package handlers

import (
	"net/http"

	"github.com/Markikie/cinema-booking/internal/middleware"
	"github.com/Markikie/cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo       *repository.UserRepository
	googleClientID string
	jwtSecret      string
}

func NewAuthHandler(userRepo *repository.UserRepository, googleClientID, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userRepo:       userRepo,
		googleClientID: googleClientID,
		jwtSecret:      jwtSecret,
	}
}

type loginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type loginResponse struct {
	Token string      `json:"token"`
	User  userSummary `json:"user"`
}

type userSummary struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token is required"})
		return
	}

	claims, err := middleware.VerifyGoogleToken(c.Request.Context(), req.IDToken, h.googleClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid google token"})
		return
	}

	user, err := h.userRepo.FindOrCreate(c.Request.Context(), claims.Subject, claims.Email, claims.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create or find user"})
		return
	}

	appToken, err := middleware.GenerateAppToken(h.jwtSecret, user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token: appToken,
		User: userSummary{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Role:  string(user.Role),
		},
	})
}
