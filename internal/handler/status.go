package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPublicStatus(c *gin.Context) {
	monitors, err := h.svc.GetPublicMonitors(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, monitors)
}
