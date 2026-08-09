package handler

import (
	"errors"
	"net/http"

	"github.com/Umesh-Tiruvalluru/monitoring/internal/models"
	"github.com/Umesh-Tiruvalluru/monitoring/internal/service"
	"github.com/gin-gonic/gin"
)

type registerResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var in models.RegisterUser
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.svc.RegisterUser(c.Request.Context(), &in)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, registerResponse{User: user, Token: token})
}

func (h *Handler) LoginUser(c *gin.Context) {
	var in models.LoginUser
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.svc.LoginUser(c.Request.Context(), &in)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCreds) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := UserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
