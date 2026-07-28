package httpapi

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"tinyschool-api/internal/academicyears"
	"tinyschool-api/internal/account"
	"tinyschool-api/internal/assignments"
	"tinyschool-api/internal/classes"
	"tinyschool-api/internal/exams"
	"tinyschool-api/internal/overview"
	"tinyschool-api/internal/schools"
	"tinyschool-api/internal/students"
	"tinyschool-api/internal/webjson"
)

func NewHandler(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		webjson.Write(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	account.Register(mux)
	overview.Register(mux)
	schools.Register(mux)
	academicyears.Register(mux)
	classes.Register(mux)
	students.Register(mux)
	assignments.Register(mux)
	exams.Register(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		webjson.Error(w, http.StatusNotFound, "not_found", "endpoint not found")
	})

	return recoverer(logger, cors(mux))
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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
				webjson.Error(w, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func NewServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
