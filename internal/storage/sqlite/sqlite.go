package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db         *sql.DB
	folderPath string
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	absPath, err := filepath.Abs(storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	dir := filepath.Dir(absPath)

	return &Storage{db: db, folderPath: dir}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

func (s *Storage) IndexFile(ctx context.Context,
	fileName string,
) (createTime, updateTime time.Time, err error) {

	return time.Now(), time.Now(), nil // TODO: delete this
}

func (s *Storage) GetFile(ctx context.Context, fileName string) ([]byte, error) {

	return []byte{}, nil // TODO: delete this
}

func (s *Storage) GetFolder() (path string) {
	return s.folderPath
}
