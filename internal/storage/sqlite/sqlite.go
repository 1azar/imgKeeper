package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	imgKeeperv1 "github.com/1azar/imgKeeper-api-contracts/gen/go/imgKeeper"
	"imgKeeper/internal/storage"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	tableName     = "fileLog"
	fileNameCol   = "fileName"
	createDateCol = "createDate"
	updateDateCol = "updateDate"
	filePathCol   = "filePath"
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

	query := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
	    %s TEXT PRIMARY KEY,
	    %s DATETIME,
	    %s DATETIME,
	    %s TEXT
	)`, tableName, fileNameCol, createDateCol, updateDateCol, filePathCol)

	_, err = db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", fn, err)
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
	fileFolder string,
) (createTime, updateTime time.Time, err error) {
	const fn = "storage.sqlite.IndexFile"

	var resCreateDate time.Time
	var resUpdateDate time.Time

	// SQL-запрос для вставки или обновления записи
	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (%s, %s, %s, %s)
		VALUES (?, COALESCE((SELECT %s FROM %s WHERE %s = ?), CURRENT_TIMESTAMP), ?, ?)
	`, tableName, fileNameCol, createDateCol, updateDateCol, filePathCol, createDateCol, tableName, fileNameCol)

	// Выполняем запрос
	_, err = s.db.Exec(query, fileName, fileName, time.Now(), filepath.Join(fileFolder, fileName))

	if err != nil {
		return resCreateDate, resUpdateDate, fmt.Errorf("%s : %w", fn, err)
	}

	queryGetRow := fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s = ?", createDateCol, updateDateCol, tableName, fileNameCol)
	rows, err := s.db.Query(queryGetRow, fileName)
	if err != nil {
		return resCreateDate, resUpdateDate, fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&resCreateDate, &resUpdateDate)
		if err != nil {
			return resCreateDate, resUpdateDate, fmt.Errorf("%s: %w", fn, err)
		}
	} else {
		return resCreateDate, resUpdateDate, fmt.Errorf("%s: %w", fn, err)
	}

	return resCreateDate, resUpdateDate, nil
}

func (s *Storage) GetFile(ctx context.Context, fileName string) ([]byte, error) {

	return []byte{}, nil // TODO: delete this
}

func (s *Storage) IsFileExist(ctx context.Context, fileName string) (ok bool, path string, err error) {
	const fn = "storage.sqlite.IsFileExist"

	var exists int
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", tableName, fileNameCol)
	err = s.db.QueryRow(query, fileName).Scan(&exists)
	if err != nil {
		return false, "", fmt.Errorf("%s: %w", fn, err)
	}

	if exists != 1 {
		return false, "", fmt.Errorf("%s: %w", fn, storage.FileDoesNotExist)
	}

	queryGetRow := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", filePathCol, tableName, fileNameCol)
	rows, err := s.db.Query(queryGetRow, fileName)
	if err != nil {
		return false, "", fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()
	var resFilePath string
	if rows.Next() {
		err = rows.Scan(&resFilePath)
		if err != nil {
			return false, "", err
		}
	}
	return true, resFilePath, nil
}

func (s *Storage) GetFolder() (path string) {
	return s.folderPath
}

// GetFileList reads all table data -> returns formatted string
func (s *Storage) GetFileList(ctx context.Context) (data string, err error) {
	const fn = "storage.sqlite.GetFileList"

	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s", fileNameCol, createDateCol, updateDateCol, tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return "none", fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	//var data string
	for rows.Next() {
		var fname, cd, ud string
		if err := rows.Scan(&fname, &cd, &ud); err != nil {
			return "none", fmt.Errorf("%s: %w", fn, err)
		}

		data = fmt.Sprintf("%s\t%s\t%s\n", fn, cd, ud)
	}

	return data, nil
}

// SendToStream send table data by chunks to stream
func (s *Storage) SendToStream(stream imgKeeperv1.ImgKeeper_ImgListServer) error {
	const fn = "storage.sqlite.SendToStream"

	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s", fileNameCol, createDateCol, updateDateCol, tableName)
	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	defer rows.Close()

	buf := make([]byte, 1024)

	for rows.Next() {
		var fname, cdate, udate string
		if err := rows.Scan(&fname, &cdate, &udate); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		data := fmt.Sprintf("%s\t%s\t%s\n", fname, cdate, udate)
		copy(buf, []byte(data))

		if err := stream.Send(&imgKeeperv1.ImgListRes{Chunk: buf}); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}

		// nullify buffer
		buf = make([]byte, 1024)

	}

	return nil
}
