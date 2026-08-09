// Package emailer delivers downtime alerts to monitor owners via the Resend API.
package emailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const ResendAPIURL = "https://api.resend.com/emails"

type AlertDetail struct {
	MonitorName string
	MonitorURL  string
	StatusCode  int
	Error       string
	CheckedAt   time.Time
}

type Emailer interface {
	SendDowntimeAlert(ctx context.Context, to string, detail AlertDetail) error
}

type ResendEmailer struct {
	APIKey    string
	FromEmail string
	HTTP      *http.Client
}

func NewResendEmailer() *ResendEmailer {
	return &ResendEmailer{
		APIKey:    os.Getenv("RESEND_API_KEY"),
		FromEmail: os.Getenv("RESEND_FROM_EMAIL"),
		HTTP:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *ResendEmailer) Enabled() bool {
	return r != nil && r.APIKey != ""
}

func (r *ResendEmailer) SendDowntimeAlert(ctx context.Context, to string, detail AlertDetail) error {
	subject := fmt.Sprintf("[Uptime Alert] %s is DOWN", detail.MonitorName)
	htmlBody := buildHTMLBody(detail)
	textBody := buildTextBody(detail)

	if !r.Enabled() {
		log.Printf("emailer: RESEND_API_KEY not set — logging downtime alert for monitor %q (%s): to=%s status=%d error=%q checkedAt=%s",
			detail.MonitorName, detail.MonitorURL, to, detail.StatusCode, detail.Error, detail.CheckedAt.Format(time.RFC3339))
		return nil
	}

	if strings.TrimSpace(r.FromEmail) == "" {
		return fmt.Errorf("emailer: RESEND_FROM_EMAIL is required when RESEND_API_KEY is set")
	}
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("emailer: recipient email is empty")
	}

	payload := map[string]any{
		"from":    r.FromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
		"text":    textBody,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("emailer: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ResendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("emailer: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := r.HTTP
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("emailer: post to resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("emailer: resend returned %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

func buildTextBody(d AlertDetail) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Monitor: %s\n", d.MonitorName)
	fmt.Fprintf(&b, "URL: %s\n", d.MonitorURL)
	fmt.Fprintf(&b, "Status: HTTP %d\n", d.StatusCode)
	if d.Error != "" {
		fmt.Fprintf(&b, "Error: %s\n", d.Error)
	}
	fmt.Fprintf(&b, "Checked At: %s\n", d.CheckedAt.Format(time.RFC3339))
	return b.String()
}

func buildHTMLBody(d AlertDetail) string {
	var b strings.Builder
	b.WriteString("<div style=\"font-family:system-ui,sans-serif;line-height:1.5\">")
	fmt.Fprintf(&b, "<h2 style=\"color:#dc2626;margin:0 0 12px 0\">%s is DOWN</h2>", htmlEscape(d.MonitorName))
	b.WriteString("<table style=\"border-collapse:collapse\">")
	fmt.Fprintf(&b, "<tr><td style=\"padding:4px 12px 4px 0;color:#6b7280\">Monitor</td><td style=\"padding:4px 0\"><strong>%s</strong></td></tr>", htmlEscape(d.MonitorName))
	fmt.Fprintf(&b, "<tr><td style=\"padding:4px 12px 4px 0;color:#6b7280\">URL</td><td style=\"padding:4px 0\"><a href=\"%s\">%s</a></td></tr>", htmlEscape(d.MonitorURL), htmlEscape(d.MonitorURL))
	fmt.Fprintf(&b, "<tr><td style=\"padding:4px 12px 4px 0;color:#6b7280\">Status Code</td><td style=\"padding:4px 0\">HTTP %d</td></tr>", d.StatusCode)
	if d.Error != "" {
		fmt.Fprintf(&b, "<tr><td style=\"padding:4px 12px 4px 0;color:#6b7280\">Error</td><td style=\"padding:4px 0\">%s</td></tr>", htmlEscape(d.Error))
	}
	fmt.Fprintf(&b, "<tr><td style=\"padding:4px 12px 4px 0;color:#6b7280\">Checked At</td><td style=\"padding:4px 0\">%s</td></tr>", d.CheckedAt.UTC().Format(time.RFC3339))
	b.WriteString("</table></div>")
	return b.String()
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
