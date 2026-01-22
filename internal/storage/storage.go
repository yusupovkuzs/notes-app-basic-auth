package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/yusupovkuzs/GoNotesappBasicAuth/internal/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

const (
	UsersTable = "users"
	NotesTable = "notes"
)

var (
	ErrAccessDenied  = errors.New("access denied")
	ErrUsernameTaken = errors.New("username taken")
)

type StoragePostgres struct {
	DB *sql.DB
}

func NewStoragePostgres(dbInfo config.PostgresConfig) (*StoragePostgres, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbInfo.Host, dbInfo.Port, dbInfo.User, dbInfo.Password, dbInfo.DBName,
	)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return &StoragePostgres{DB: db}, nil
}

func RunMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}
