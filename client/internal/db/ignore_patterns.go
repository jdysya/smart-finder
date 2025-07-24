package db

import (
	"database/sql"
	"log"
	"strings"
)

// GetIgnoredPatterns retrieves all ignored patterns from the database.
func GetIgnoredPatterns(dbConn *sql.DB) ([]string, error) {
	rows, err := dbConn.Query("SELECT pattern FROM ignored_patterns")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []string
	for rows.Next() {
		var pattern string
		if err := rows.Scan(&pattern); err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

// UpdateIgnoredPatterns clears existing patterns and saves a new list of patterns.
func UpdateIgnoredPatterns(dbConn *sql.DB, patternsText string) error {
	patterns := strings.Split(strings.ReplaceAll(patternsText, "\r\n", "\n"), "\n")

	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}

	// Clear existing patterns
	_, err = tx.Exec("DELETE FROM ignored_patterns")
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert new patterns
	stmt, err := tx.Prepare("INSERT INTO ignored_patterns (pattern) VALUES (?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, p := range patterns {
		pattern := strings.TrimSpace(p)
		if pattern != "" {
			_, err := stmt.Exec(pattern)
			if err != nil {
				log.Printf("Failed to insert ignore pattern '%s': %v", pattern, err)
			}
		}
	}

	return tx.Commit()
}
