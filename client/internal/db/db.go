package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	// 开启 WAL 模式，提高并发能力
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS files (
        md5 TEXT PRIMARY KEY,
        path TEXT,
        filename TEXT,
        size INTEGER,
        modified_at DATETIME
    );
	CREATE TABLE IF NOT EXISTS monitored_directories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL UNIQUE
	);
    `
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("create table failed: %w", err)
	}
	return db, nil
}
