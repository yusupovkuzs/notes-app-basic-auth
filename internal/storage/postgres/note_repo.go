package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/models"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/storage"
)

type NoteRepository interface {
	CreateNote(note models.Note) (int, error)
	GetAllNotes(userId, limit, offset int, sort string) ([]models.NoteDTO, error)
	GetNote(userId, noteId int) (models.NoteDTO, error)
	UpdateNote(userId, noteId int, note models.Note) (int, error)
	DeleteNote(userId, noteId int) error
}

type NoteRepoPostgres struct {
	db *sql.DB
}

func NewNoteRepoPostgres(db *sql.DB) *NoteRepoPostgres {
	return &NoteRepoPostgres{db: db}
}

func (r *NoteRepoPostgres) CreateNote(n models.Note) (int, error) {
	const op = "storage.postgres.GetUser"

	var id int

	query := fmt.Sprintf(
		"INSERT INTO %s (user_id, title, content) VALUES ($1, $2, $3) RETURNING id",
		storage.NotesTable,
	)
	row := r.db.QueryRow(query, n.UserID, n.Title, n.Content)
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *NoteRepoPostgres) GetAllNotes(userId, limit, offset int, sort string) ([]models.NoteDTO, error) {
	const op = "storage.postgres.GetAllNotes"

	var notes []models.NoteDTO

	query := fmt.Sprintf(
		`SELECT id, title, content, created_at, updated_at 
				FROM %s 
				WHERE user_id = $1 
				ORDER BY created_at %s
				LIMIT $2 OFFSET $3 `,
		storage.NotesTable, sort,
	)

	rows, err := r.db.Query(query, userId, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var n models.NoteDTO

		err = rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return notes, nil
}

func (r *NoteRepoPostgres) GetNote(userId, noteId int) (models.Note, error) {
	const op = "storage.postgres.GetNote"

	var n models.Note
	n.ID = noteId

	query := fmt.Sprintf(
		"SELECT user_id, title, content, created_at, updated_at FROM %s WHERE id = $1",
		storage.NotesTable,
	)

	row := r.db.QueryRow(query, noteId)
	if err := row.Scan(&n.UserID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return models.Note{}, fmt.Errorf("%s: %w", op, err)
	}
	if n.UserID != userId {
		return models.Note{}, fmt.Errorf("%s: note is not owned by %d", op, userId)
	}

	return n, nil
}

func (r *NoteRepoPostgres) UpdateNote(userId, noteId int, note models.UpdateNoteInput) error {
	const op = "storage.postgres.UpdateNote"

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if note.Title != nil {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, *note.Title)
		argId++
	}

	if note.Content != nil {
		setValues = append(setValues, fmt.Sprintf("content=$%d", argId))
		args = append(args, *note.Content)
		argId++
	}

	setValues = append(setValues, "updated_at=now()")

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(
		`UPDATE %s
		 SET %s
		 WHERE id = $%d AND user_id = $%d`,
		storage.NotesTable,
		setQuery,
		argId,
		argId+1,
	)

	args = append(args, noteId, userId)
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *NoteRepoPostgres) DeleteNote(userId, noteId int) error {
	const op = "storage.postgres.Delete"

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE id = $1 AND user_id = $2 RETURNING id",
		storage.NotesTable,
	)
	var deletedID int
	err := r.db.QueryRow(query, noteId, userId).Scan(&deletedID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
