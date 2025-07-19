package indexer

import (
	"database/sql"
	"log"
)

func RemoveFile(db *sql.DB, path string) {
	log.Println("从索引中删除文件:", path)
	_, err := db.Exec("DELETE FROM files WHERE path = ?", path)
	if err != nil {
		log.Println("删除文件索引失败:", err)
	}
}
