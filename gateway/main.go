package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"smart-finder/gateway/internal/templates"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Kodbox   KodboxConfig   `mapstructure:"kodbox"`
	Error    ErrorConfig    `mapstructure:"error"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

type ServerConfig struct {
	Port   int    `mapstructure:"port"`
	Domain string `mapstructure:"domain"`
}

type KodboxConfig struct {
	Domain string `mapstructure:"domain"`
}

type ErrorConfig struct {
	NotFoundPage string `mapstructure:"not_found_page"`
}

type MD5Gateway struct {
	config    *Config
	db        *sql.DB
	templates *templates.Templates
}

func main() {
	// 加载配置
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("无法解析配置: %v", err)
	}

	// 连接数据库
	db, err := connectDatabase(&config.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 初始化模板
	tmpl, err := templates.New()
	if err != nil {
		log.Fatalf("模板初始化失败: %v", err)
	}

	// 创建网关实例
	gateway := &MD5Gateway{
		config:    &config,
		db:        db,
		templates: tmpl,
	}

	// 设置路由
	router := mux.NewRouter()

	// 静态文件服务 - 使用embed.FS
	router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", getPublicFileServer()))

	router.HandleFunc("/md5", gateway.handleMD5Query).Methods("GET")
	router.HandleFunc("/api/md5", gateway.handleMD5API).Methods("GET")

	// 启动服务器
	serverAddr := ":" + strconv.Itoa(config.Server.Port)
	log.Printf("MD5网关服务器启动在端口 %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, router))
}

func connectDatabase(dbConfig *DatabaseConfig) (*sql.DB, error) {
	dsn := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" +
		dbConfig.Host + ":" + strconv.Itoa(dbConfig.Port) + ")/" +
		dbConfig.Database + "?charset=" + dbConfig.Charset + "&parseTime=True&loc=Local"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func (g *MD5Gateway) handleMD5Query(w http.ResponseWriter, r *http.Request) {
	// 获取MD5哈希值
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "缺少MD5哈希参数", http.StatusBadRequest)
		return
	}

	// 验证MD5格式（32位十六进制字符）
	if len(hash) != 32 {
		http.Error(w, "无效的MD5哈希格式", http.StatusBadRequest)
		return
	}

	// 使用模板渲染页面
	data := templates.TemplateData{
		Hash:         hash,
		ServerDomain: g.config.Server.Domain,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := g.templates.RenderMD5Page(w, data); err != nil {
		log.Printf("模板渲染失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
		return
	}
}

// 处理API MD5查询（服务端处理）
func (g *MD5Gateway) handleMD5API(w http.ResponseWriter, r *http.Request) {
	// 获取MD5哈希值
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "缺少MD5哈希参数", http.StatusBadRequest)
		return
	}

	// 验证MD5格式（32位十六进制字符）
	if len(hash) != 32 {
		http.Error(w, "无效的MD5哈希格式", http.StatusBadRequest)
		return
	}

	// 查询文件ID
	fileID, err := g.getFileIDByMD5(hash)
	if err != nil {
		log.Printf("查询文件ID失败: %v", err)
		http.Error(w, "数据库查询错误", http.StatusInternalServerError)
		return
	}

	if fileID == 0 {
		// 未找到文件，重定向到错误页面
		http.Redirect(w, r, g.config.Error.NotFoundPage, http.StatusFound)
		return
	}

	// 查询源ID
	sourceID, err := g.getSourceIDByFileID(fileID)
	if err != nil {
		log.Printf("查询源ID失败: %v", err)
		http.Error(w, "数据库查询错误", http.StatusInternalServerError)
		return
	}

	if sourceID == 0 {
		// 未找到源，重定向到错误页面
		http.Redirect(w, r, g.config.Error.NotFoundPage, http.StatusFound)
		return
	}

	// 重定向到客户端
	redirectURL := g.config.Kodbox.Domain + "/#explorer&sidf=" + strconv.Itoa(sourceID)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (g *MD5Gateway) getFileIDByMD5(hash string) (int, error) {
	var fileID int
	query := "SELECT fileID FROM io_file WHERE hashMD5 = ? LIMIT 1"
	err := g.db.QueryRow(query, hash).Scan(&fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // 未找到记录
		}
		return 0, err
	}
	return fileID, nil
}

func (g *MD5Gateway) getSourceIDByFileID(fileID int) (int, error) {
	var sourceID int
	query := "SELECT sourceID FROM io_source WHERE fileId = ? LIMIT 1"
	err := g.db.QueryRow(query, fileID).Scan(&sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // 未找到记录
		}
		return 0, err
	}
	return sourceID, nil
}
