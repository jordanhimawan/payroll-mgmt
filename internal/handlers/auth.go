package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jordanhimawan/payroll-mgmt/internal/models"
	"github.com/jordanhimawan/payroll-mgmt/internal/services"
	"github.com/jordanhimawan/payroll-mgmt/pkg/response"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Username == "" || req.Password == "" {
		response.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	loginResp, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if appErr, ok := err.(*services.AppError); ok {
			response.Error(w, appErr.Message, appErr.Code)
			return
		}
		response.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response.JSON(w, loginResp, http.StatusOK)
}
