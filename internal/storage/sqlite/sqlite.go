package sqlite

import (
	"database/sql"
	"fmt"
	"imageProcessor/internal/models"
	"sync"
)

type StorageSqlite struct {
	db *sql.DB
	mu sync.RWMutex
}

func New(storagePath string) (*StorageSqlite, error) {
	const op = "sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s; error opening sqlite3 storage: %v", op, err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Соединения не устаревают
	db.SetConnMaxIdleTime(0) // Бездействующие соединения не закрываются

	return &StorageSqlite{db: db}, nil
}

// SetMetadata function is used to store any image
// into local storage - /uploads;
// And this function create image metadata in sqlite
func (s *StorageSqlite) SetMetadata(metadata *models.ImageMetadata) (id int, err error) {
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

func (s *StorageSqlite) GetImageMetadata(id int) (*models.ImageMetadata, error) {
	const op = "sqlite.GetImageMetadata"

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

func (s *StorageSqlite) DeleteImage(id int) error {
	const op = "sqlite.DeleteImage"

	res, err := s.db.Exec(`DELETE FROM images WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("%s,%w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s,%w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("do not delete string by id=%d; %s,%w", id, op, err)
	}
	return nil
}
