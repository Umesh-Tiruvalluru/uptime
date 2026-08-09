package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) CreateMonitor(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	var in models.CreateMonitorRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.svc.CreateMonitor(c.Request.Context(), userID, &in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetAllMonitors(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	monitors, err := h.svc.GetAllMonitors(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, monitors)
}

func (h *Handler) GetMonitorByID(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}

	m, err := h.svc.GetMonitorByID(c.Request.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMonitorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) UpdateMonitor(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}

	var in models.UpdateMonitorRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.svc.UpdateMonitor(c.Request.Context(), userID, id, &in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMonitorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) DeleteMonitor(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}

	if err := h.svc.DeleteMonitor(c.Request.Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, service.ErrMonitorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// GetMonitorHistory returns the most recent check results for a monitor
// in chronological order.
func (h *Handler) GetMonitorHistory(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor id"})
		return
	}

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 200"})
			return
		}
		limit = parsed
	}

	history, err := h.svc.GetCheckHistory(c.Request.Context(), userID, id, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMonitorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, history)
}
