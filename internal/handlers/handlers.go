package handlers

import "github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage/postgres"

type Handlers struct {
	noteRepo *postgres.NoteRepoPostgres
	userRepo *postgres.UserRepoPostgres
}

func NewHandlers(noteRepo *postgres.NoteRepoPostgres, userRepo *postgres.UserRepoPostgres) *Handlers {
	return &Handlers{
		noteRepo: noteRepo,
		userRepo: userRepo,
	}
}
