package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"embed"
	"io/fs"
	"smart-finder/client/internal/db"
	"smart-finder/client/internal/indexer"
	"smart-finder/client/internal/utils"

	"github.com/fsnotify/fsnotify"
)

var (
	monitoredDirs   = make([]string, 0)
	monitoredDirsMu sync.RWMutex
	dbConn          *sql.DB
	watcher         *fsnotify.Watcher
)

// CORS中间件
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Check-Request")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

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

	// 从数据库加载监控目录
	rows, err := dbConn.Query("SELECT path FROM monitored_directories")
	if err != nil {
		log.Fatal("加载监控目录失败:", err)
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			monitoredDirsMu.Lock()
			monitoredDirs = append(monitoredDirs, path)
			monitoredDirsMu.Unlock()
		}
	}

	// 初始化 watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("创建文件监控器失败:", err)
	}
	defer watcher.Close()

	// 将监控目录及其子目录添加到 watcher
	monitoredDirsMu.RLock()
	for _, dir := range monitoredDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("访问路径 %q 失败: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					log.Printf("添加监控目录 %q 失败: %v\n", path, err)
				}
			}
			return nil
		})
	}
	monitoredDirsMu.RUnlock()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("检测到文件变动:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					indexer.Scanner(dbConn, event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					indexer.RemoveFile(dbConn, event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("文件监控器错误:", err)
			}
		}
	}()

	// 静态文件服务（使用 embed.FS）
	webRoot, _ := fs.Sub(webFS, "web")
	http.Handle("/", http.FileServer(http.FS(webRoot)))

	// API 路由
	http.HandleFunc("/api/directories", directoriesHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/health", corsMiddleware(healthHandler))
	http.HandleFunc("/api/path2url", path2urlHandler)
	http.HandleFunc("/api/files", filesHandler)
	http.HandleFunc("/api/md5", apiMD5FileHandler)
	http.HandleFunc("/api/ignore-patterns", ignorePatternsHandler)

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

// 新增：通过md5返回本地文件内容
func apiMD5FileHandler(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if len(hash) != 32 {
		http.Error(w, "参数错误，缺少或错误的md5", 400)
		return
	}

	var filePath, fileName string
	err := dbConn.QueryRow("SELECT path, filename FROM files WHERE md5 = ?", hash).Scan(&filePath, &fileName)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "数据库错误", 500)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "文件无法打开", 500)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "文件信息获取失败", 500)
		return
	}

	// 自动推断Content-Type
	ext := filepath.Ext(fileName)
	var contentType string
	if ext == ".md" || ext == ".markdown" {
		contentType = "text/markdown; charset=utf-8"
	} else {
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			// 读一部分内容推断
			buf := make([]byte, 512)
			n, _ := f.Read(buf)
			contentType = http.DetectContentType(buf[:n])
			f.Seek(0, io.SeekStart)
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+fileName+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

	// 支持Range请求（视频/大文件友好）
	http.ServeContent(w, r, fileName, fi.ModTime(), f)
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

		// 添加到数据库
		db.UpdateMonitoredDir(dbConn, req.Path, "add")

		// 添加到 watcher
		filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Printf("访问路径 %q 失败: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					log.Printf("添加监控目录 %q 失败: %v\n", path, err)
				}
			}
			return nil
		})

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

		// 从数据库删除
		db.UpdateMonitoredDir(dbConn, req.Path, "remove")

		// 从 watcher 中移除
		err := watcher.Remove(req.Path)
		if err != nil {
			log.Println("移除监控目录失败:", err)
		}
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

func ignorePatternsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		patterns, err := db.GetIgnoredPatterns(dbConn)
		if err != nil {
			http.Error(w, "Failed to get ignored patterns", 500)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		json.NewEncoder(w).Encode(patterns)
	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", 400)
			return
		}
		if err := db.UpdateIgnoredPatterns(dbConn, string(body)); err != nil {
			http.Error(w, "Failed to update ignored patterns", 500)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Unsupported method", 405)
	}
}

// 获取所有已索引文件，支持分页
func filesHandler(w http.ResponseWriter, r *http.Request) {
	// 解析分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	search := r.URL.Query().Get("search")
	page := 1
	pageSize := 20
	var err error
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			pageSize = 20
		}
	}
	offset := (page - 1) * pageSize

	// 构造SQL和参数
	var (
		where string
		args  []interface{}
	)
	if search != "" {
		where = "WHERE filename LIKE ? OR path LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	// 查询总数
	totalSql := "SELECT COUNT(*) FROM files " + where
	var total int
	err = dbConn.QueryRow(totalSql, args...).Scan(&total)
	if err != nil {
		http.Error(w, "数据库错误", 500)
		return
	}

	// 查询分页数据
	dataSql := "SELECT md5, path, filename, size, modified_at FROM files " + where + " ORDER BY modified_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	rows, err := dbConn.Query(dataSql, args...)
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"total": total,
	})
}
