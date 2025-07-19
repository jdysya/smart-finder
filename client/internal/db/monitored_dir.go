package db

import (
	"database/sql"
	"log"
)

func UpdateMonitoredDir(dbConn *sql.DB, path string, action string) {
	switch action {
	case "add":
		_, err := dbConn.Exec("INSERT INTO monitored_directories (path) VALUES (?)", path)
		if err != nil {
			log.Println("添加监控目录到数据库失败:", err)
		}
	case "remove":
		tx, err := dbConn.Begin()
		if err != nil {
			log.Println("开启事务失败:", err)
			return
		}
		// 删除目录本身
		_, err = tx.Exec("DELETE FROM monitored_directories WHERE path = ?", path)
		if err != nil {
			log.Println("从数据库删除监控目录失败:", err)
			tx.Rollback()
			return
		}
		// 删除该目录下的所有文件索引
		_, err = tx.Exec("DELETE FROM files WHERE path LIKE ?", path+"%")
		if err != nil {
			log.Println("删除文件索引失败:", err)
			tx.Rollback()
			return
		}
		err = tx.Commit()
		if err != nil {
			log.Println("提交事务失败:", err)
		}
	}
}
