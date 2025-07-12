package constants

const (
	// 客户端配置
	ClientPort = 8964
	ClientHost = "127.0.0.1"
	
	// 服务端配置
	DefaultServerPort = 8080
	
	// 数据库配置
	SQLiteDBPath = "data/md5fs.db"
	
	// API路径
	HealthEndpoint = "/api/health"
	MD5Endpoint    = "/md5"
	
	// 请求头
	CheckRequestHeader = "X-Check-Request"
	
	// 版本信息
	Version = "1.0.0"
	
	// 超时配置
	ClientTimeout = 3000 // 毫秒
	FileCheckTimeout = 5000 // 毫秒
)
