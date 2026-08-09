package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateMonitor now seeds next_check_at = NOW() so a freshly created monitor is
// immediately eligible for the scheduler to pick up.
func (r *Repository) CreateMonitor(ctx context.Context, m *models.Monitor) error {
	const q = `
        INSERT INTO monitors (user_id, name, url, interval_seconds, next_check_at)
        VALUES ($1, $2, $3, $4, NOW())
        RETURNING id, created_at, updated_at, next_check_at, last_status`
	return r.db.QueryRow(ctx, q, m.UserID, m.Name, m.URL, m.IntervalSecs).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt, &m.NextCheckAt, &m.LastStatus)
}

func (r *Repository) GetMonitorsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Monitor, error) {
	const q = `
		SELECT m.id, m.user_id, m.name, m.url, m.interval_seconds, m.next_check_at, m.last_checked_at,
		       m.last_status, latest.response_time, m.created_at, m.updated_at
		FROM monitors m
		LEFT JOIN LATERAL (
			SELECT response_time FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC LIMIT 1
		) latest ON TRUE
		WHERE m.user_id = $1 ORDER BY m.created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]models.Monitor, 0)
	for rows.Next() {
		m, err := scanMonitor(rows)
		if err != nil {
			return nil, err
		}
		monitors = append(monitors, *m)
	}
	return monitors, rows.Err()
}

func (r *Repository) GetMonitorByID(ctx context.Context, id uuid.UUID) (*models.Monitor, error) {
	const q = `
		SELECT m.id, m.user_id, m.name, m.url, m.interval_seconds, m.next_check_at, m.last_checked_at,
		       m.last_status, latest.response_time, m.created_at, m.updated_at
		FROM monitors m
		LEFT JOIN LATERAL (
			SELECT response_time FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC LIMIT 1
		) latest ON TRUE
		WHERE m.id = $1`
	m := &models.Monitor{}
	err := r.db.QueryRow(ctx, q, id).
		Scan(&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSecs, &m.NextCheckAt,
			&m.LastCheckedAt, &m.LastStatus, &m.LastResponseMs, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) UpdateMonitor(ctx context.Context, id uuid.UUID, in *models.UpdateMonitorRequest) (*models.Monitor, error) {
	// Interval change reschedules the next check.
	const q = `
        UPDATE monitors SET
            name = COALESCE($2, name),
            url = COALESCE($3, url),
            interval_seconds = COALESCE($4, interval_seconds),
            updated_at = NOW(),
            next_check_at = CASE WHEN $4 IS NOT NULL THEN NOW() ELSE next_check_at END
        WHERE id = $1
		RETURNING id, user_id, name, url, interval_seconds, next_check_at, last_checked_at,
		          last_status,
		          (SELECT response_time FROM check_results WHERE monitor_id = monitors.id ORDER BY checked_at DESC LIMIT 1),
		          created_at, updated_at`
	m := &models.Monitor{}
	err := r.db.QueryRow(ctx, q, id, in.Name, in.URL, in.IntervalSecs).
		Scan(&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSecs, &m.NextCheckAt,
			&m.LastCheckedAt, &m.LastStatus, &m.LastResponseMs, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) DeleteMonitor(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetDueMonitors atomically picks monitors whose next_check_at <= NOW()
// and bumps next_check_at by their interval so other scheduler instances
// don't double-publish. Uses FOR UPDATE SKIP LOCKED for safety.
func (r *Repository) GetDueMonitors(ctx context.Context, now time.Time, limit int) ([]models.Monitor, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const selectQ = `
		SELECT m.id, m.user_id, m.name, m.url, m.interval_seconds, m.next_check_at, m.last_checked_at,
		       m.last_status, latest.response_time, m.created_at, m.updated_at
		FROM monitors m
		LEFT JOIN LATERAL (
			SELECT response_time FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC LIMIT 1
		) latest ON TRUE
		WHERE m.next_check_at <= $1
		ORDER BY m.next_check_at ASC
        LIMIT $2
		FOR UPDATE OF m SKIP LOCKED`
	rows, err := tx.Query(ctx, selectQ, now, limit)
	if err != nil {
		return nil, err
	}
	claimed := make([]models.Monitor, 0)
	for rows.Next() {
		m, err := scanMonitor(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		claimed = append(claimed, *m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(claimed) == 0 {
		return nil, tx.Commit(ctx)
	}

	// Advance next_check_at so the next scheduler tick won't re-claim these.
	const advanceQ = `
        UPDATE monitors SET next_check_at = NOW() + (interval_seconds || ' seconds')::interval
        WHERE id = ANY($1)`
	ids := make([]uuid.UUID, len(claimed))
	for i, m := range claimed {
		ids[i] = m.ID
	}
	if _, err := tx.Exec(ctx, advanceQ, ids); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return claimed, nil
}

// RecordCheck persists a single check result and updates the monitor's denormalized
// last_status / last_checked_at in one transaction.
func (r *Repository) RecordCheck(ctx context.Context, res *models.CheckResult) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const insertCheckQ = `
		INSERT INTO check_results
            (monitor_id, response_time, status_code, error_message, status, checked_at)
        VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, insertCheckQ,
		res.MonitorID, res.ResponseMs, res.StatusCode, res.ErrorMessage, string(res.Status), res.CheckedAt,
	); err != nil {
		return err
	}

	const updateMonitorQ = `
        UPDATE monitors SET
            last_status = $2,
			last_checked_at = $3,
			updated_at = NOW()
        WHERE id = $1`
	if _, err := tx.Exec(ctx, updateMonitorQ, res.MonitorID, string(res.Status), res.CheckedAt); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetPublicMonitors(ctx context.Context) ([]models.PublicMonitor, error) {
	const q = `
		SELECT m.id, m.name, m.url, m.last_status, m.last_checked_at, latest.response_time
		FROM monitors m
		LEFT JOIN LATERAL (
			SELECT response_time FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC LIMIT 1
		) latest ON TRUE
		ORDER BY m.name ASC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]models.PublicMonitor, 0)
	for rows.Next() {
		var monitor models.PublicMonitor
		if err := rows.Scan(&monitor.ID, &monitor.Name, &monitor.URL, &monitor.LastStatus, &monitor.LastCheckedAt, &monitor.LastResponseMs); err != nil {
			return nil, err
		}
		monitors = append(monitors, monitor)
	}
	return monitors, rows.Err()
}

func (r *Repository) GetCheckHistory(ctx context.Context, monitorID uuid.UUID, limit int) ([]models.CheckResult, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT monitor_id, status, status_code, response_time, error_message, checked_at
		FROM check_results
		WHERE monitor_id = $1
		ORDER BY checked_at DESC
		LIMIT $2`
	rows, err := r.db.Query(ctx, q, monitorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]models.CheckResult, 0)
	for rows.Next() {
		var res models.CheckResult
		var statusStr string
		if err := rows.Scan(&res.MonitorID, &statusStr, &res.StatusCode, &res.ResponseMs, &res.ErrorMessage, &res.CheckedAt); err != nil {
			return nil, err
		}
		res.Status = models.MonitorStatus(statusStr)
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reverse slice to present chronological order (oldest -> newest)
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	return results, nil
}


// scanMonitor is a small helper because rows.Scan and QueryRow.Scan can't share
// a closure-bound destination set in Go without an interface adapter.
func scanMonitor(rows pgx.Row) (*models.Monitor, error) {
	m := &models.Monitor{}
	if err := rows.Scan(
		&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSecs, &m.NextCheckAt,
		&m.LastCheckedAt, &m.LastStatus, &m.LastResponseMs, &m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return m, nil
}
