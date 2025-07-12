# MD5文件定位系统 部署文档

## 系统架构

- **客户端**: 运行在用户PC上，提供本地文件索引和定位功能
- **服务端**: 部署在Linux服务器上，提供网关和数据库查询功能

## 构建

### 使用构建脚本
```bash
./scripts/build.sh
```

### 手动构建
```bash
# 构建客户端 (Windows)
cd client
GOOS=windows GOARCH=amd64 go build -o ../build/md5-fs-client-windows.exe .

# 构建服务端 (Linux)
cd gateway
GOOS=linux GOARCH=amd64 go build -o ../build/md5-fs-gateway-linux .
```

## 部署

### 客户端部署
1. 将构建好的客户端可执行文件分发给用户
2. 用户运行客户端程序
3. 客户端会在127.0.0.1:8964端口启动

### 服务端部署
1. 将构建好的服务端可执行文件上传到Linux服务器
2. 配置数据库连接信息
3. 启动服务端程序
4. 服务端会在配置的端口启动

### 使用部署脚本
```bash
./scripts/deploy.sh
```

## 配置

### 服务端配置
编辑 `gateway/config/config.yaml`:
```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  database: "kodbox"
  charset: "utf8mb4"

server:
  port: 8080
  domain: "http://your-domain.com"

error:
  not_found_page: "/error/not-found.html"
```

### 客户端配置
客户端会自动创建SQLite数据库文件在 `client/data/md5fs.db`

## 系统要求

### 客户端
- Windows 10+ / macOS 10.14+ / Linux
- 至少100MB可用磁盘空间

### 服务端
- Linux (推荐 Ubuntu 20.04+)
- MySQL 5.7+ 或 MariaDB 10.3+
- 至少1GB可用内存
- 至少500MB可用磁盘空间

## 监控和维护

### 日志
- 客户端日志: 控制台输出
- 服务端日志: 控制台输出

### 健康检查
- 客户端: `http://127.0.0.1:8964/api/health`
- 服务端: 检查进程状态

## 故障排除

### 常见问题
1. **客户端无法启动**: 检查端口8964是否被占用
2. **服务端连接数据库失败**: 检查数据库配置和网络连接
3. **跨域问题**: 确保客户端CORS配置正确
