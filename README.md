# MD5文件定位系统

这是一个基于MD5哈希值的文件定位系统，支持本地客户端和服务端双重处理机制。

## 功能特性

- **智能路由**: 自动检测本地客户端状态，优先使用本地客户端处理
- **双重处理**: 本地客户端不可用时自动切换到服务端处理
- **文件检查**: 本地客户端会检查文件是否存在，不存在时转发到服务端
- **健康检查**: 提供客户端健康状态检查接口

## 项目结构

```
smart-finder/
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

### 本地构建

#### 使用构建脚本（推荐）
```bash
# 构建所有组件
./scripts/build.sh

# 只构建客户端
./scripts/build.sh client

# 只构建服务端
./scripts/build.sh gateway

# 测试交叉编译
./scripts/build.sh cross

# 运行测试
./scripts/build.sh test
```

#### 手动构建
```bash
# 构建客户端
cd client
go build -ldflags="-s -w" -o smart-finder-client .

# 构建服务端
cd gateway
go build -ldflags="-s -w" -o smart-finder-gateway .
```

### GitHub Actions 自动构建

项目配置了GitHub Actions工作流，支持：

1. **自动测试构建**：每次推送代码时自动测试构建
2. **多平台构建**：支持Windows、Linux、macOS平台
3. **Release打包**：发布Release时自动生成所有平台的安装包

详细说明请查看 [GitHub Actions 文档](docs/github-actions.md)

### 构建产物

构建完成后会生成以下文件：
- `smart-finder-client-linux-amd64` - Linux客户端 (兼容CentOS 7+)
- `smart-finder-client-windows-amd64.exe` - Windows客户端
- `smart-finder-client-darwin-amd64` - macOS Intel客户端
- `smart-finder-client-darwin-arm64` - macOS Apple Silicon客户端
- `smart-finder-gateway-linux-amd64` - Linux服务端 (兼容CentOS 7+)
- `smart-finder-gateway-windows-amd64.exe` - Windows服务端
- `smart-finder-gateway-darwin-amd64` - macOS Intel服务端
- `smart-finder-gateway-darwin-arm64` - macOS Apple Silicon服务端

### CentOS 7 兼容性

项目已针对CentOS 7进行了优化：

- **静态链接**: 使用静态链接避免动态库依赖问题
- **架构兼容**: 设置 `GOAMD64=v1` 确保与旧版本CPU兼容
- **纯Go实现**: 使用 `osusergo,netgo` 标签确保纯Go实现

详细说明请查看 [CentOS 7 兼容性文档](docs/centos7-compatibility.md)

#### 本地构建CentOS 7兼容版本
```bash
# 使用专用构建脚本
./scripts/build-centos7.sh

# 或手动构建
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOAMD64=v1

cd client
go build -ldflags="-s -w -extldflags=-static" -tags="osusergo,netgo" -o smart-finder-client-linux-amd64-centos7 .
cd ../gateway
go build -ldflags="-s -w -extldflags=-static" -tags="osusergo,netgo" -o smart-finder-gateway-linux-amd64-centos7 .
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