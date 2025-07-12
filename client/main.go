package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"embed"
	"io/fs"
	"smart-finder/client/internal/db"
	"smart-finder/client/internal/indexer"
	"smart-finder/client/internal/utils"
)

var (
	monitoredDirs   = make([]string, 0)
	monitoredDirsMu sync.RWMutex
	dbConn          *sql.DB
)

// CORS中间件
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Check-Request")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

//go:embed web/*
var webFS embed.FS

func main() {
	var err error
	dbConn, err = db.InitDB("data/md5fs.db")
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 静态文件服务（使用 embed.FS）
	webRoot, _ := fs.Sub(webFS, "web")
	http.Handle("/", http.FileServer(http.FS(webRoot)))

	// API 路由
	http.HandleFunc("/api/directories", directoriesHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/health", corsMiddleware(healthHandler))
	http.HandleFunc("/api/path2url", path2urlHandler)
	http.HandleFunc("/api/files", filesHandler)

	// md5 文件定位路由
	http.HandleFunc("/md5", corsMiddleware(md5Handler))

	port := 8964
	log.Printf("服务启动: http://127.0.0.1:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil))
}

// 处理 /md5 路由
func md5Handler(w http.ResponseWriter, r *http.Request) {
	// 期望路径格式 /md5?hash=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	hash := r.URL.Query().Get("hash")
	if len(hash) != 32 {
		http.Error(w, "参数错误，缺少或错误的md5", 400)
		return
	}

	// 检查是否是检查请求
	isCheckRequest := r.Header.Get("X-Check-Request") == "true"

	var filePath string
	err := dbConn.QueryRow("SELECT path FROM files WHERE md5 = ?", hash).Scan(&filePath)
	if err == sql.ErrNoRows {
		if isCheckRequest {
			// 如果是检查请求，返回404状态但不显示错误页面
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "数据库错误", 500)
		return
	}

	// 如果是检查请求，只返回成功状态
	if isCheckRequest {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 正常处理：在文件管理器中定位文件
	err = utils.RevealInExplorer(filePath)
	if err != nil {
		http.Error(w, "打开文件失败", 500)
		return
	}
	w.Write([]byte("已在文件管理器中定位文件"))
}

// 监控目录API
func directoriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		monitoredDirsMu.RLock()
		defer monitoredDirsMu.RUnlock()
		json.NewEncoder(w).Encode(monitoredDirs)
	case "POST":
		var req struct{ Path string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, "参数错误", 400)
			return
		}
		monitoredDirsMu.Lock()
		monitoredDirs = append(monitoredDirs, req.Path)
		monitoredDirsMu.Unlock()
		go indexer.Scanner(dbConn, req.Path)
		w.WriteHeader(201)
	case "DELETE":
		var req struct{ Path string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, "参数错误", 400)
			return
		}
		monitoredDirsMu.Lock()
		for i, p := range monitoredDirs {
			if p == req.Path {
				monitoredDirs = append(monitoredDirs[:i], monitoredDirs[i+1:]...)
				break
			}
		}
		monitoredDirsMu.Unlock()
		w.WriteHeader(204)
	default:
		http.Error(w, "不支持的方法", 405)
	}
}

// 状态API
func statusHandler(w http.ResponseWriter, r *http.Request) {
	var count int
	dbConn.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
	status := map[string]interface{}{
		"indexing":      indexer.Indexing,
		"fileCount":     count,
		"indexingTotal": indexer.IndexingTotal,
		"indexingDone":  indexer.IndexingDone,
	}
	json.NewEncoder(w).Encode(status)
}

// 健康检查API
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	json.NewEncoder(w).Encode(health)
}

// 路径转url API
func path2urlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只支持POST", 405)
		return
	}
	var req struct{ Path string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
		http.Error(w, "参数错误", 400)
		return
	}
	var md5 string
	err := dbConn.QueryRow("SELECT md5 FROM files WHERE path = ?", req.Path).Scan(&md5)
	if err == sql.ErrNoRows {
		http.Error(w, "路径未被索引", 404)
		return
	} else if err != nil {
		http.Error(w, "数据库错误", 500)
		return
	}
	url := "/md5?hash=" + md5
	json.NewEncoder(w).Encode(map[string]string{
		"md5": md5,
		"url": url,
	})
}

// 获取所有已索引文件
func filesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := dbConn.Query("SELECT md5, path, filename, size, modified_at FROM files")
	if err != nil {
		http.Error(w, "数据库错误", 500)
		return
	}
	defer rows.Close()
	type FileInfo struct {
		MD5        string `json:"md5"`
		Path       string `json:"path"`
		Filename   string `json:"filename"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modified_at"`
	}
	var files []FileInfo
	for rows.Next() {
		var f FileInfo
		if err := rows.Scan(&f.MD5, &f.Path, &f.Filename, &f.Size, &f.ModifiedAt); err == nil {
			files = append(files, f)
		}
	}
	json.NewEncoder(w).Encode(files)
}
