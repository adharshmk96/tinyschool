package httpdelivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"tinyschool-api/internal/service"
)

const maxRequestBody = 1 << 20

type itemResponse struct {
	Data any `json:"data"`
}

type collectionResponse struct {
	Data any      `json:"data"`
	Meta pageMeta `json:"meta"`
}

type pageMeta struct {
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeItem(w http.ResponseWriter, status int, value any) {
	writeJSON(w, status, itemResponse{Data: value})
}

func writeCollection(w http.ResponseWriter, value any, total, page, pageSize int) {
	writeJSON(w, http.StatusOK, collectionResponse{
		Data: value,
		Meta: pageMeta{Total: total, Page: page, PageSize: pageSize},
	})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message}})
}

func writeServiceError(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var serviceError *service.Error
	message := http.StatusText(http.StatusInternalServerError)
	if errors.As(err, &serviceError) && serviceError.Message != "" {
		message = serviceError.Message
	}

	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", message)
	case errors.Is(err, service.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized", message)
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", message)
	case errors.Is(err, service.ErrConflict):
		writeError(w, http.StatusConflict, "conflict", message)
	default:
		logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return fmt.Errorf("request body is required")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return fmt.Errorf("%s", err)
		default:
			return fmt.Errorf("malformed JSON: %w", err)
		}
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON value")
	}
	return nil
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("request panic", "method", r.Method, "path", r.URL.Path, "error", recovered, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
