package httpdelivery

import (
	"net/http"

	"tinyschool-api/internal/dto"
)

func (h *Handler) listClasses(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListClasses(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getClass(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetClassFiltered(r.Context(), r.PathValue("id"), r.URL.Query().Get("grade"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createClass(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.ClassRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateClass(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateClass(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.UpdateClassRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateClass(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteClass(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteClass(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listStudents(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListStudents(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getStudent(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetStudent(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createStudent(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.StudentRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateStudent(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateStudent(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.UpdateStudentRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateStudent(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteStudent(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteStudent(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addBehaviour(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.BehaviourRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.AddBehaviour(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) deleteBehaviour(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteBehaviour(r.Context(), r.PathValue("id"), r.PathValue("logId")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addNote(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.NoteRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.AddNote(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) deleteNote(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteNote(r.Context(), r.PathValue("id"), r.PathValue("logId")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
