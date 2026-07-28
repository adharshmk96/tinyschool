package httpdelivery

import (
	"net/http"

	"tinyschool-api/internal/dto"
)

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.Overview(r.Context())
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) listSchools(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListSchools(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getSchool(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetSchool(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createSchool(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.SchoolRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateSchool(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateSchool(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.SchoolRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateSchool(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteSchool(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteSchool(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listAcademicYears(w http.ResponseWriter, r *http.Request) {
	options, ok := readListOptions(w, r)
	if !ok {
		return
	}
	result, err := h.app.ListAcademicYears(r.Context(), options)
	if err != nil {
		writeListError(h.logger, w, r, err)
		return
	}
	writeCollection(w, result.Items, result.Total, result.Page, result.PageSize)
}

func (h *Handler) getAcademicYear(w http.ResponseWriter, r *http.Request) {
	result, err := h.app.GetAcademicYear(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) createAcademicYear(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.AcademicYearRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.CreateAcademicYear(r.Context(), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusCreated, result)
}

func (h *Handler) updateAcademicYear(w http.ResponseWriter, r *http.Request) {
	input, ok := decodeBody[dto.AcademicYearRequest](w, r)
	if !ok {
		return
	}
	result, err := h.app.UpdateAcademicYear(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, result)
}

func (h *Handler) deleteAcademicYear(w http.ResponseWriter, r *http.Request) {
	if err := h.app.DeleteAcademicYear(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
