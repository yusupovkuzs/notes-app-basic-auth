package middleware

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func CheckUserAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := r.Context().Value("user_id").(int)
		paramID := chi.URLParam(r, "id")

		if strconv.Itoa(ctxUserID) != paramID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
