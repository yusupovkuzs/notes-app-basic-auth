package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/config"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/handlers"
	mw "github.com/yusupovkuzs/GoNotesappBasicAuth/internal/middleware"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage/postgres"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/logger"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/pkg/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// config
	cfg := config.MustLoad()
	// logger
	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting Notes App", slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled")

	// storage
	database, err := storage.NewStoragePostgres(cfg.Postgres)
	if err != nil {
		log.Error("DB connection failed: %w", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Database connected successfully")

	// migrations
	if err = storage.RunMigrations(database.DB); err != nil {
		log.Error("Migrations failed: %w", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Migrations completed")

	// router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(mw.RequestLogger(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	noteRepo := postgres.NewNoteRepoPostgres(database.DB)
	userRepo := postgres.NewUserRepoPostgres(database.DB)
	handler := handlers.NewHandlers(noteRepo, userRepo)

	router.Post("/users", handler.Register(log))
	router.Route("/users/{id}", func(r chi.Router) {
		r.Use(handler.BasicAuth(log))
		r.Use(mw.CheckUserAccess)

		r.Post("/notes", handler.CreateNote(log))
		r.Get("/notes", handler.GetAllNotes(log))
		r.Get("/notes/{note_id}", handler.GetNote(log))
		r.Put("/notes/{note_id}", handler.UpdateNote(log))
		r.Delete("/notes/{note_id}", handler.DeleteNote(log))
	})

	// start server
	log.Info("starting server", slog.String("address", cfg.HttpServer.Address))
	srv := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.ReadTimeout,
		WriteTimeout: cfg.HttpServer.WriteTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to stop server")
		return
	}

	log.Info("server stopped")
}
