package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xchwan/simple-web-framework/framework"
)

// UserHandler 負責處理會員相關的 HTTP 請求。
type UserHandler struct {
	service *UserService
}

// NewUserHandler 建立一個 UserHandler。
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// ===== Request / Response DTO =====

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type renameRequest struct {
	NewName string `json:"newName"`
}

type userResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type loginResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

// ===== Handlers =====

// Register 處理 POST /api/users。
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := framework.ParseRequest(r, &req); err != nil {
		framework.RespondText(w, r, http.StatusBadRequest, "Registration's format incorrect.")
		return
	}
	u, err := h.service.Register(req.Email, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailDuplicate):
			framework.RespondText(w, r, http.StatusBadRequest, "Duplicate email")
		default:
			framework.RespondText(w, r, http.StatusBadRequest, "Registration's format incorrect.")
		}
		return
	}
	framework.Respond(w, r, http.StatusOK, userResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	})
}

// Login 處理 POST /api/users/login。
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := framework.ParseRequest(r, &req); err != nil {
		framework.RespondText(w, r, http.StatusBadRequest, "Login's format incorrect.")
		return
	}
	u, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrCredentialsInvalid):
			framework.RespondText(w, r, http.StatusBadRequest, "Credentials Invalid")
		default:
			framework.RespondText(w, r, http.StatusBadRequest, "Login's format incorrect.")
		}
		return
	}
	framework.Respond(w, r, http.StatusOK, loginResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Token: u.Token,
	})
}

// UpdateName 處理 PATCH /api/users/{userId}。
func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	caller, err := h.service.Authenticate(token)
	if err != nil {
		framework.RespondText(w, r, http.StatusUnauthorized, "Can't authenticate who you are.")
		return
	}

	targetID, err := strconv.Atoi(framework.PathParam(r, "userId"))
	if err != nil {
		framework.RespondText(w, r, http.StatusBadRequest, "Name's format invalid.")
		return
	}

	var req renameRequest
	if err := framework.ParseRequest(r, &req); err != nil {
		framework.RespondText(w, r, http.StatusBadRequest, "Name's format invalid.")
		return
	}

	if err := h.service.UpdateName(caller.ID, targetID, req.NewName); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			framework.RespondText(w, r, http.StatusForbidden, "Forbidden")
		default:
			framework.RespondText(w, r, http.StatusBadRequest, "Name's format invalid.")
		}
		return
	}
	framework.Respond(w, r, http.StatusNoContent, nil)
}

// SearchUsers 處理 GET /api/users。
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if _, err := h.service.Authenticate(token); err != nil {
		framework.RespondText(w, r, http.StatusUnauthorized, "Can't authenticate who you are.")
		return
	}

	keyword := r.URL.Query().Get("keyword")
	users := h.service.SearchUsers(keyword)

	result := make([]userResponse, len(users))
	for i, u := range users {
		result[i] = userResponse{ID: u.ID, Email: u.Email, Name: u.Name}
	}
	framework.Respond(w, r, http.StatusOK, result)
}

// ===== 私有輔助函式 =====

// extractToken 從 Authorization: Bearer <token> 標頭取出 token。
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
