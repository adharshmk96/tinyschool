package httpdelivery

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"tinyschool-api/internal/dataio"
)

// exportData streams the caller's whole workspace as a spreadsheet: an XLSX
// workbook by default, or a ZIP of CSV files with ?format=csv.
func (h *Handler) exportData(w http.ResponseWriter, r *http.Request) {
	format, err := dataio.ParseFormat(r.URL.Query().Get("format"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	dataset, err := h.app.ExportData(r.Context())
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	body, err := dataio.Encode(dataset, format)
	if err != nil {
		h.logger.Error("export failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "the export could not be created")
		return
	}
	filename := fmt.Sprintf("tinyschool-export-%s.%s", time.Now().UTC().Format("2006-01-02"), format.Extension())
	w.Header().Set("Content-Type", format.ContentType())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(body); err != nil {
		h.logger.Warn("export write failed", "error", err)
	}
}

// importData replaces the caller's workspace with an uploaded spreadsheet. The
// upload is multipart so a browser can post the file directly.
func (h *Handler) importData(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, dataio.MaxUploadBytes)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "the file could not be read; it may be larger than 25 MB")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "attach the export file as the \"file\" field")
		return
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(io.LimitReader(file, dataio.MaxUploadBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_upload", "the file could not be read")
		return
	}
	dataset, err := dataio.Decode(header.Filename, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_file", err.Error())
		return
	}
	summary, err := h.app.ImportData(r.Context(), dataset)
	if err != nil {
		writeServiceError(h.logger, w, r, err)
		return
	}
	writeItem(w, http.StatusOK, summary)
}
