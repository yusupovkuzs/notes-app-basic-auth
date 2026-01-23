package handlers

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/models"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/logger/sl"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handlers) BasicAuth(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Basic ") {
				log.Error("Authorization header missing")
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
			if err != nil {
				log.Error("Error decoding authorization header", slog.String("error", err.Error()))
				http.Error(w, "Invalid auth header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(string(payload), ":", 2)
			if len(parts) != 2 {
				log.Error("Invalid authorization header", slog.String("payload", string(payload)))
				http.Error(w, "Invalid auth header", http.StatusUnauthorized)
				return
			}

			username, password := parts[0], parts[1]
			user, err := h.userRepo.GetUser(username, password)
			if err != nil {
				log.Error("Error checking user", slog.String("error", err.Error()))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			log.Info("User authenticated", slog.Int("userID", user.ID))
			ctx := context.WithValue(r.Context(), "user_id", user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handlers) Register(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.Register"

		var input models.User

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		err := render.DecodeJSON(r.Body, &input)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("request body decoded successfully", slog.Any("request", input))

		if input.Username == "" || input.Password == "" {
			log.Error("invalid username or password")
			response.RespondError(w, http.StatusBadRequest, "invalid username or password")
			return
		}

		id, err := h.userRepo.CreateUser(input)
		if errors.Is(err, storage.ErrUsernameTaken) {
			log.Error("username is already taken", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, "username is already taken")
			return
		}
		if err != nil {
			log.Error("failed to create user", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("user created successfully", slog.Int("id", id))
		response.RespondJSON(w, http.StatusCreated, map[string]interface{}{
			"id": id,
		})
	}
}
