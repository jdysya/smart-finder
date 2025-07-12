# MD5文件定位系统

这是一个基于MD5哈希值的文件定位系统，支持本地客户端和服务端双重处理机制。

## 功能特性

- **智能路由**: 自动检测本地客户端状态，优先使用本地客户端处理
- **双重处理**: 本地客户端不可用时自动切换到服务端处理
- **文件检查**: 本地客户端会检查文件是否存在，不存在时转发到服务端
- **健康检查**: 提供客户端健康状态检查接口

## 项目结构

```
md5-fs/
├── README.md
├── go.work                    # Go工作区文件
├── .gitignore
├── docs/                      # 文档
│   ├── api.md
│   └── deployment.md
├── scripts/                   # 构建和部署脚本
│   ├── build.sh
│   └── deploy.sh
├── client/                    # 客户端
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── internal/
│   ├── data/
│   └── web/
├── gateway/                   # 服务端
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── config/
│   └── public/
└── shared/                    # 共享代码
    ├── go.mod
    ├── types/
    ├── utils/
    └── constants/
```

## 系统架构

### 服务端 (gateway)
- 端口: 8080 (可配置)
- 提供 `/md5?hash={md5}` 接口
- 返回智能处理页面，自动检测客户端状态
- 提供 `/api/md5?hash={md5}` 接口用于服务端处理

### 客户端 (client)
- 端口: 8964 (固定)
- 提供本地文件索引和定位功能
- 健康检查接口: `/api/health`
- 文件检查接口: `/md5?hash={md5}` (带 `X-Check-Request: true` 头)

### 共享模块 (shared)
- 提供公共类型定义、工具函数和常量
- 被客户端和服务端共同使用

## 使用流程

1. **访问MD5查询页面**: `http://localhost:8080/md5?hash={md5}`
2. **自动检测**: 页面自动检查本地客户端 (127.0.0.1:8964) 状态
3. **智能处理**:
   - 如果客户端可用且文件存在 → 重定向到本地客户端
   - 如果客户端可用但文件不存在 → 使用服务端处理
   - 如果客户端不可用 → 使用服务端处理

## 启动方法

### 使用Go工作区（推荐）
```bash
# 启动服务端
go run ./gateway

# 启动客户端
go run ./client
```

### 传统方式
```bash
# 启动服务端
cd gateway
go run main.go

# 启动客户端
cd client
go run main.go
```

## 测试

访问测试页面: `http://localhost:8080/public/test.html`

## API接口

### 服务端接口

#### GET /md5?hash={md5}
返回智能处理页面，自动检测客户端状态并决定处理方式。

#### GET /api/md5?hash={md5}
服务端直接处理MD5查询，重定向到文件管理器。

### 客户端接口

#### GET /api/health
健康检查接口，返回客户端状态信息。

#### GET /md5?hash={md5}
文件定位接口，在文件管理器中定位文件。

#### GET /md5?hash={md5} (带 X-Check-Request: true 头)
文件存在性检查接口，只检查文件是否存在，不执行定位操作。

## 构建和部署

### 构建所有平台版本
```bash
./scripts/build.sh
```

这将生成以下文件：
- `build/md5-fs-client-windows.exe` - Windows客户端
- `build/md5-fs-client-darwin` - macOS客户端  
- `build/md5-fs-client-linux` - Linux客户端
- `build/md5-fs-gateway-linux` - Linux服务端
- `build/md5-fs-gateway-windows.exe` - Windows服务端

### 部署
```bash
./scripts/deploy.sh
```

## 配置

### 服务端配置 (gateway/config/config.yaml)
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

## 技术实现

### 前端检测逻辑
1. 使用 `fetch()` 调用客户端健康检查接口
2. 如果客户端可用，检查文件是否存在
3. 根据检查结果决定处理方式

### 客户端检查机制
- 使用 `X-Check-Request` 请求头区分检查请求和正常请求
- 检查请求只返回状态码，不执行文件定位操作
- 配置了CORS支持，允许跨域请求

### CORS配置
客户端已配置CORS中间件，支持以下功能：
- 允许所有来源的跨域请求 (`Access-Control-Allow-Origin: *`)
- 支持GET、POST、OPTIONS方法
- 允许自定义请求头 `X-Check-Request`
- 自动处理预检请求 (OPTIONS)

### 错误处理
- 客户端不可用时自动降级到服务端
- 文件不存在时提供友好的错误页面
- 网络超时和连接错误的优雅处理

## 注意事项

1. 确保客户端在 127.0.0.1:8964 端口运行
2. 服务端需要正确配置数据库连接
3. 客户端需要正确配置监控目录
4. 客户端已配置CORS支持，允许跨域请求
5. 如果遇到跨域问题，请确保客户端正在运行且端口正确 