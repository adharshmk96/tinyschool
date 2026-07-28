package httpdelivery

import (
	"net/http"
	"time"

	"tinyschool-api/internal/dto"
)

func (h *Handler) registerProtectedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/me", h.me)
	mux.HandleFunc("PATCH /api/v1/me", h.updateMe)
	mux.HandleFunc("PUT /api/v1/me/password", h.updatePassword)
	mux.HandleFunc("GET /api/v1/overview", h.overview)
	mux.HandleFunc("GET /api/v1/schools", h.listSchools)
	mux.HandleFunc("POST /api/v1/schools", h.createSchool)
	mux.HandleFunc("GET /api/v1/schools/{id}", h.getSchool)
	mux.HandleFunc("PATCH /api/v1/schools/{id}", h.updateSchool)
	mux.HandleFunc("DELETE /api/v1/schools/{id}", h.deleteSchool)
	mux.HandleFunc("GET /api/v1/academic-years", h.listAcademicYears)
	mux.HandleFunc("POST /api/v1/academic-years", h.createAcademicYear)
	mux.HandleFunc("GET /api/v1/academic-years/{id}", h.getAcademicYear)
	mux.HandleFunc("PATCH /api/v1/academic-years/{id}", h.updateAcademicYear)
	mux.HandleFunc("DELETE /api/v1/academic-years/{id}", h.deleteAcademicYear)
	mux.HandleFunc("GET /api/v1/classes", h.listClasses)
	mux.HandleFunc("POST /api/v1/classes", h.createClass)
	mux.HandleFunc("GET /api/v1/classes/{id}", h.getClass)
	mux.HandleFunc("PATCH /api/v1/classes/{id}", h.updateClass)
	mux.HandleFunc("DELETE /api/v1/classes/{id}", h.deleteClass)
	mux.HandleFunc("GET /api/v1/students", h.listStudents)
	mux.HandleFunc("POST /api/v1/students", h.createStudent)
	mux.HandleFunc("GET /api/v1/students/{id}", h.getStudent)
	mux.HandleFunc("PATCH /api/v1/students/{id}", h.updateStudent)
	mux.HandleFunc("DELETE /api/v1/students/{id}", h.deleteStudent)
	mux.HandleFunc("POST /api/v1/students/{id}/behaviour", h.addBehaviour)
	mux.HandleFunc("DELETE /api/v1/students/{id}/behaviour/{logId}", h.deleteBehaviour)
	mux.HandleFunc("POST /api/v1/students/{id}/notes", h.addNote)
	mux.HandleFunc("DELETE /api/v1/students/{id}/notes/{logId}", h.deleteNote)
	mux.HandleFunc("GET /api/v1/assignments", h.listAssignments)
	mux.HandleFunc("POST /api/v1/assignments", h.createAssignment)
	mux.HandleFunc("GET /api/v1/assignments/{id}", h.getAssignment)
	mux.HandleFunc("PATCH /api/v1/assignments/{id}", h.updateAssignment)
	mux.HandleFunc("DELETE /api/v1/assignments/{id}", h.deleteAssignment)
	mux.HandleFunc("PUT /api/v1/assignments/{id}/scores/{studentId}", h.setAssignmentScore)
	mux.HandleFunc("DELETE /api/v1/assignments/{id}/scores/{studentId}", h.clearAssignmentScore)
	mux.HandleFunc("GET /api/v1/exams", h.listExams)
	mux.HandleFunc("POST /api/v1/exams", h.createExam)
	mux.HandleFunc("GET /api/v1/exams/{id}", h.getExam)
	mux.HandleFunc("PATCH /api/v1/exams/{id}", h.updateExam)
	mux.HandleFunc("DELETE /api/v1/exams/{id}", h.deleteExam)
	mux.HandleFunc("PUT /api/v1/exams/{id}/scores/{studentId}", h.setExamScore)
	mux.HandleFunc("DELETE /api/v1/exams/{id}/scores/{studentId}", h.clearExamScore)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.RegisterRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.Register(r.Context(), input)
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	setSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusCreated, result.User)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.LoginRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.Login(r.Context(), input)
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}
	setSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusOK, result.User)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.Refresh(r.Context(), requestToken(r))
	if err != nil {
		clearSessionCookie(w)
		writeServiceError(h.logger, w, r, err)
		return
	}
	setSessionCookie(w, result.Token, result.ExpiresAt)
	writeItem(w, http.StatusOK, result.User)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.app.Logout(r.Context(), requestToken(r)); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, err := h.app.Me(r.Context(), requestIdentity(r).UserID)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, user)
}

func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.UpdateUserRequest](w, r)
	if !ok {
		return
	}
	user, err := h.app.UpdateMe(r.Context(), requestIdentity(r).UserID, input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, user)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.PasswordRequest](w, r)
	if !ok {
		return
	}
	identity := requestIdentity(r)
	if err := h.app.UpdatePassword(r.Context(), identity.UserID, identity.SessionID, input); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	writeServiceError(h.logger, w, r, err)
}

func setSessionCookie(w http.ResponseWriter, token, expiresAt string) {
	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		expires = time.Now().Add(24 * time.Hour)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
		MaxAge:   max(1, int(time.Until(expires).Seconds())),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: "", Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0),
	})
}
