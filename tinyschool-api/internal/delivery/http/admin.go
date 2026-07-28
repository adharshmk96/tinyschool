package httpdelivery

import (
	"context"
	"net/http"
	"time"

	"tinyschool-api/internal/dto"
)

// adminSessionCookieName is deliberately separate from the school session
// cookie so an administrator and a teacher can be signed in side by side.
const adminSessionCookieName = "tinyschool_admin_session"

func (h *Handler) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/admin/status", h.adminStatus)
	mux.HandleFunc("POST /api/v1/admin/setup", h.adminSetup)
	mux.HandleFunc("POST /api/v1/admin/login", h.adminLogin)
	mux.HandleFunc("POST /api/v1/admin/refresh", h.adminRefresh)
	mux.HandleFunc("POST /api/v1/admin/logout", h.adminLogout)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/admin/me", h.adminMe)
	protected.HandleFunc("GET /api/v1/admin/users", h.adminListUsers)
	protected.HandleFunc("POST /api/v1/admin/admins", h.adminAddAdmin)
	protected.HandleFunc("DELETE /api/v1/admin/users/{id}", h.adminDeleteUser)
	protected.HandleFunc("POST /api/v1/admin/users/{id}/block", h.adminBlockUser)
	protected.HandleFunc("POST /api/v1/admin/users/{id}/unblock", h.adminUnblockUser)
	mux.Handle("/api/v1/admin/", h.authenticateAdmin(protected))
}

type adminContextKey struct{}

func (h *Handler) authenticateAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admin, err := h.app.AuthenticateAdmin(r.Context(), adminRequestToken(r))
		if err != nil {
			writeServiceError(h.logger, w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminContextKey{}, admin)))
	})
}

func requestAdmin(r *http.Request) dto.AdminUser {
	value, _ := r.Context().Value(adminContextKey{}).(dto.AdminUser)
	return value
}

func adminRequestToken(r *http.Request) string {
	cookie, err := r.Cookie(adminSessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (h *Handler) adminStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.app.AdminStatus(r.Context())
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, status)
}

func (h *Handler) adminSetup(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.AdminSetupRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateAdmin(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	setAdminSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusCreated, result.User)
}

func (h *Handler) adminLogin(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.LoginRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.AdminLogin(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	setAdminSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusOK, result.User)
}

func (h *Handler) adminRefresh(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.AdminRefresh(r.Context(), adminRequestToken(r))
	if err != nil {
		clearAdminSessionCookie(w)
		writeServiceError(h.logger, w, r, err)
		return
	}
	setAdminSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusOK, result.User)
}

func (h *Handler) adminLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.app.Logout(r.Context(), adminRequestToken(r)); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	clearAdminSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminMe(w http.ResponseWriter, r *http.Request) {
	writeItem(w, http.StatusOK, requestAdmin(r))
}

func (h *Handler) adminListUsers(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListUsers(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

// adminAddAdmin sits behind authenticateAdmin: adding administrators is only
// possible for someone who already holds an admin session.
func (h *Handler) adminAddAdmin(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.AdminSetupRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.AddAdmin(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteUser(r.Context(), requestAdmin(r).ID, r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminBlockUser(w http.ResponseWriter, r *http.Request) {
	h.setUserBlocked(w, r, true)
}

func (h *Handler) adminUnblockUser(w http.ResponseWriter, r *http.Request) {
	h.setUserBlocked(w, r, false)
}

func (h *Handler) setUserBlocked(w http.ResponseWriter, r *http.Request, blocked bool) {
	result, err := h.app.SetUserBlocked(r.Context(), r.PathValue("id"), blocked)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func setAdminSessionCookie(w http.ResponseWriter, token, expiresAt string) {
	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		expires = time.Now().Add(24 * time.Hour)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
		MaxAge:   max(1, int(time.Until(expires).Seconds())),
	})
}

func clearAdminSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: adminSessionCookieName, Value: "", Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0),
	})
}
