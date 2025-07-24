package indexer

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"smart-finder/client/internal/core"
	"smart-finder/client/internal/db"
)

var (
	Indexing      bool
	IndexingTotal int
	IndexingDone  int
)

func Scanner(dbConn *sql.DB, root string) error {
	Indexing = true
	defer func() { Indexing = false }()

	ignorePatterns, err := db.GetIgnoredPatterns(dbConn)
	if err != nil {
		log.Printf("Error getting ignored patterns: %v", err)
		// Decide if you want to continue without ignore patterns or return the error
	}

	// First, count the total number of files to be indexed.
	total := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that can't be accessed
		}

		// Check against ignore patterns
		for _, pattern := range ignorePatterns {
			matched, _ := filepath.Match(pattern, info.Name())
			if matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() {
			total++
		}
		return nil
	})
	IndexingTotal = total
	IndexingDone = 0

	// Formal indexing
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing a file or directory: %s, error: %v, skipping", path, err)
			return nil
		}

		// Check against ignore patterns
		for _, pattern := range ignorePatterns {
			matched, _ := filepath.Match(pattern, info.Name())
			if matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil // Skip directories themselves
		}

		md5sum, err := CalculateMD5(path)
		if err != nil {
			log.Printf("Error calculating MD5 for: %s, error: %v, skipping", path, err)
			return nil
		}

		fileIndex := core.FileIndex{
			MD5:        md5sum,
			Path:       path,
			Filename:   info.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		}

		// Insert or update
		_, err = dbConn.Exec(`
            INSERT INTO files (md5, path, filename, size, modified_at)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(md5) DO UPDATE SET
                path=excluded.path,
                filename=excluded.filename,
                size=excluded.size,
                modified_at=excluded.modified_at
        `, fileIndex.MD5, fileIndex.Path, fileIndex.Filename, fileIndex.Size, fileIndex.ModifiedAt)
		if err != nil {
			log.Printf("Failed to index file: %s, error: %v", path, err)
		}

		IndexingDone++
		return nil
	})
}

