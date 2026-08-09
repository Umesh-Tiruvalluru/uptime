package handler

import (
	"net/http"
	"strings"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ctxUserKey = "userID"

// AuthMiddleware verifies the Bearer token and stores the user ID on the context.
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		userID, err := h.svc.ParseToken(parts[1])
		if err != nil || userID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(ctxUserKey, userID)
		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ctxUserKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

var _ = service.ErrInvalidCreds
