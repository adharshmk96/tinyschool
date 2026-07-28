package account

import (
	"net/http"

	"tinyschool-api/internal/webjson"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/me", func(w http.ResponseWriter, _ *http.Request) {
		webjson.Write(w, http.StatusOK, map[string]any{"data": User{
			ID: "usr_001", Name: "Alex Morgan", Email: "alex@tinyschool.local",
		}})
	})
}
