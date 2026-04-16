package user

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/xchwan/simple-web-framework/framework"
)


// UserHandler 負責處理會員相關的 HTTP 請求。
// service 在每個 request 時從 container 動態取得，以支援 HttpRequestScope。
type UserHandler struct{}

// NewUserHandler 建立一個 UserHandler。
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// service 從 request context 的 container 取得 UserService。
func (h *UserHandler) service(r *http.Request) *UserService {
	return framework.Get[*UserService](r, "userService")
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
		framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Registration's format incorrect."})
		return
	}
	u, err := h.service(r).Register(req.Email, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailDuplicate):
			framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Duplicate email"})
		default:
			framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Registration's format incorrect."})
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
		framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Login's format incorrect."})
		return
	}
	u, err := h.service(r).Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrCredentialsInvalid):
			framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Credentials Invalid"})
		default:
			framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Login's format incorrect."})
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
	caller, err := h.service(r).Authenticate(token)
	if err != nil {
		framework.Respond(w, r, http.StatusUnauthorized, framework.ErrorBody{"Can't authenticate who you are."})
		return
	}

	targetID, err := strconv.Atoi(framework.PathParam(r, "userId"))
	if err != nil {
		framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Name's format invalid."})
		return
	}

	var req renameRequest
	if err := framework.ParseRequest(r, &req); err != nil {
		framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Name's format invalid."})
		return
	}

	if err := h.service(r).UpdateName(caller.ID, targetID, req.NewName); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			framework.Respond(w, r, http.StatusForbidden, framework.ErrorBody{"Forbidden"})
		default:
			framework.Respond(w, r, http.StatusBadRequest, framework.ErrorBody{"Name's format invalid."})
		}
		return
	}
	framework.Respond(w, r, http.StatusNoContent, nil)
}

// SearchUsers 處理 GET /api/users。
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if _, err := h.service(r).Authenticate(token); err != nil {
		framework.Respond(w, r, http.StatusUnauthorized, framework.ErrorBody{"Can't authenticate who you are."})
		return
	}

	keyword := r.URL.Query().Get("keyword")
	users := h.service(r).SearchUsers(keyword)

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
