package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Service struct {
	db *sql.DB
}

var service Service

// Connects to the Database and returns the *sql.DB
func Connect() (Service, error) {
	// Creates a new connection and return
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		return Service{}, err
	}

	// Verify the connection by pinging the DB
	err = db.Ping()
	if err != nil {
		return Service{}, err
	}

	service := Service{db: db}

	return service, nil
}

func (s *Service) Close() error {
	return s.db.Close()
}

// Creates the table in database, if only not exists
func (s *Service) CreateTable() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS movies(
		 	id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE,
			released_year TEXT,
			rating TEXT,
			Summary TEXT,
			directors TEXT,  -- Will store JSON array as text
			cast TEXT       -- Will store JSON array as text
		);
	`
	_, err := s.db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) InsertMovie(
	name string,
	released string,
	rating string,
	summary string,
	directors []string,
	cast []string,
) error {
	// Convert slices to JSON
	directorsJSON, err := json.Marshal(directors)
	if err != nil {
		return fmt.Errorf("error marshaling directors: %w", err)
	}

	castJSON, err := json.Marshal(cast)
	if err != nil {
		return fmt.Errorf("error marshaling cast: %w", err)
	}

	q := "INSERT OR IGNORE INTO movies  (name, released_year, rating, summary, directors, cast) VALUES (?, ?, ?, ?, ?, ?);"
	insert, err := s.db.Prepare(q)
	if err != nil {
		return fmt.Errorf("error preparing insert statement: %w", err)
	}
	defer insert.Close()
	res, err := insert.Exec(name, released, rating, summary, string(directorsJSON), string(castJSON))
	if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}

	// Get the auto-generated ID if needed
	_, err = res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	return nil
}
