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
func (s *Storage) SetMetadata(metadata *models.ImageMetadata) (id int, err error) {
	const op = "sqlite.UploadImage"

	row := s.db.QueryRow(`
	INSERT INTO images(original_filename, original_path, mime_type, file_size, status, action)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id;
	`, metadata.OriginalFilename, metadata.OriginalPath, metadata.MimeType, metadata.FileSize, metadata.Status, metadata.Action)

	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s %w", op, err)
	}
	return
}

func (s *Storage) DownloadImage(id int) (*models.ImageMetadata, error) {
	const op = "sqlite.DownloadImage"

	var metadata models.ImageMetadata
	row := s.db.QueryRow(`
	SELECT original_path, original_path, mime_type, file_size, status, action FROM images
	WHERE id = $1;
	`, id)

	err := row.Scan(&metadata.OriginalFilename, &metadata.OriginalPath, &metadata.MimeType, &metadata.FileSize, &metadata.Status, &metadata.Action)
	if err != nil {
		return nil, fmt.Errorf("%s,%w", op, err)
	}

	return &metadata, nil
}

func (s *Storage) DeleteImage(id int) error {
	const op = "sqlite.DeleteImage"
	return nil
}
