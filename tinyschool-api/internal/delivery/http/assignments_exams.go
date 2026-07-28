package httpdelivery

import (
	"net/http"

	"tinyschool-api/internal/dto"
)

func (h *Handler) listAssignments(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListAssignments(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getAssignment(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetAssignment(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createAssignment(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.AssignmentRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateAssignment(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateAssignment(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.UpdateAssignmentRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateAssignment(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteAssignment(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setAssignmentScore(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.ScoreRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.SetAssignmentScore(r.Context(), r.PathValue("id"), r.PathValue("studentId"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) clearAssignmentScore(w http.ResponseWriter, r *http.Request) {
	if err := h.app.ClearAssignmentScore(r.Context(), r.PathValue("id"), r.PathValue("studentId")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listExams(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListExams(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getExam(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetExam(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createExam(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.ExamRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateExam(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateExam(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.UpdateExamRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateExam(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteExam(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteExam(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setExamScore(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.ScoreRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.SetExamScore(r.Context(), r.PathValue("id"), r.PathValue("studentId"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) clearExamScore(w http.ResponseWriter, r *http.Request) {
	if err := h.app.ClearExamScore(r.Context(), r.PathValue("id"), r.PathValue("studentId")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
