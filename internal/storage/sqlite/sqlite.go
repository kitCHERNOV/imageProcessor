package sqlite

import (
	"database/sql"
	"fmt"
	"imageProcessor/internal/models"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s; error opening sqlite3 storage: %v", op, err)
	}

	return &Storage{db: db}, nil
}

// SetMetadata function is used to store any image
// into local storage - /uploads;
// And this function create image metadata in sqlite
func (s *Storage) SetMetadata(metadata *models.ImageMetadata) (int, error) {
	const op = "sqlite.UploadImage"

	s.db.QueryRow(`
	INSERT INTO 
	`)
	return nil
}

func (s *Storage) DownloadImage() error {

	return nil
}

func (s *Storage) DeleteImage() error {
	const op = "sqlite.DeleteImage"
	return nil
}
