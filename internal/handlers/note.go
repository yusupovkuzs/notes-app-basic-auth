package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/models"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/logger/sl"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func (h *Handlers) CreateNote(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.CreateNote"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var input models.Note
		err := render.DecodeJSON(r.Body, &input)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("request body decoded successfully", slog.Any("input", input))

		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to convert user id to int", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		input.UserID = userId
		id, err := h.noteRepo.CreateNote(input)
		if err != nil {
			log.Error("failed to create note", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("note created successfully", slog.Int("id", id))
		response.RespondJSON(w, http.StatusCreated, map[string]interface{}{
			"status": "OK",
			"userId": userId,
			"noteId": id,
		})
	}
}

func (h *Handlers) GetAllNotes(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetAllNotes"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		q := r.URL.Query()
		limit := 10
		offset := 0
		sort := "asc"
		if v := q.Get("limit"); v != "" {
			if l, err := strconv.Atoi(v); err == nil {
				limit = l
			}
		}

		if v := q.Get("offset"); v != "" {
			if o, err := strconv.Atoi(v); err == nil {
				offset = o
			}
		}

		if v := q.Get("sort"); v == "asc" || v == "desc" {
			sort = v
		}

		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to get id", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("user id found", slog.Any("userId", userId))

		notes, err := h.noteRepo.GetAllNotes(userId, limit, offset, sort)
		if err != nil {
			log.Error("failed to get notes", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("notes found", slog.Any("notes", notes))
		response.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"userID": userId,
			"notes":  notes,
		})
	}
}

func (h *Handlers) GetNote(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.GetNote"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		noteId := chi.URLParam(r, "note_id")
		if noteId == "" {
			log.Info("no note id provided")
			response.RespondError(w, http.StatusBadRequest, "no note id provided")
			return
		}

		noteID, err := strconv.Atoi(noteId)
		if err != nil {
			log.Error("failed to convert note id to int", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("note id found", slog.Any("noteId", noteID))

		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to get user id", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("user id found", slog.Any("userId", userId))

		note, err := h.noteRepo.GetNote(userId, noteID)
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("note not found", sl.Err(err))
			response.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, storage.ErrAccessDenied) {
			log.Info("access denied", sl.Err(err))
			response.RespondError(w, http.StatusForbidden, err.Error())
			return
		}
		if err != nil {
			log.Error("failed to get note", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("note found", slog.Any("note", note))
		response.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"userID": userId,
			"note":   note,
		})
	}
}

func (h *Handlers) UpdateNote(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.UpdateNote"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		noteId := chi.URLParam(r, "note_id")
		if noteId == "" {
			log.Info("no note id provided")
			response.RespondError(w, http.StatusBadRequest, "no note id provided")
			return
		}

		noteID, err := strconv.Atoi(noteId)
		if err != nil {
			log.Error("failed to convert note id to int", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("note id found", slog.Any("noteId", noteID))

		var input models.UpdateNoteInput
		if err = render.DecodeJSON(r.Body, &input); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("request body decoded successfully", slog.Any("input", input))

		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to get user id", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("user id found", slog.Any("userId", userId))

		err = h.noteRepo.UpdateNote(userId, noteID, input)
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("note not found", sl.Err(err))
			response.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, storage.ErrAccessDenied) {
			log.Info("access denied", sl.Err(err))
			response.RespondError(w, http.StatusForbidden, err.Error())
			return
		}
		if err != nil {
			log.Error("failed to update note", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("note updated", slog.Any("note", input))
		response.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"userID": userId,
			"noteID": noteID,
		})
	}
}

func (h *Handlers) DeleteNote(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.UpdateNote"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		noteId := chi.URLParam(r, "note_id")
		if noteId == "" {
			log.Info("no note id provided")
			response.RespondError(w, http.StatusBadRequest, "no note id provided")
			return
		}

		noteID, err := strconv.Atoi(noteId)
		if err != nil {
			log.Error("failed to convert note id to int", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("note id found", slog.Any("noteId", noteId))

		userId, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Error("failed to get user id", sl.Err(err))
			response.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Info("user id found", slog.Any("userId", userId))

		err = h.noteRepo.DeleteNote(userId, noteID)
		if errors.Is(err, sql.ErrNoRows) {
			log.Error("note not found", sl.Err(err))
			response.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, storage.ErrAccessDenied) {
			log.Info("access denied", sl.Err(err))
			response.RespondError(w, http.StatusForbidden, err.Error())
		}
		if err != nil {
			log.Error("failed to delete note", sl.Err(err))
			response.RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		log.Info("note deleted", slog.Any("note", noteId))
		response.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "OK",
			"userID": userId,
			"noteID": noteID,
		})
	}
}
