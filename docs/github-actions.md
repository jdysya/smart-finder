# GitHub Actions 构建说明

本项目包含两个GitHub Actions工作流文件，用于自动化构建和打包客户端和服务端。

## 工作流文件

### 1. `.github/workflows/build.yml` - 完整构建工作流

这个工作流用于：
- 在推送到 `main` 或 `develop` 分支时触发
- 在创建Pull Request时触发
- 在发布Release时触发

**功能：**
- 分别构建客户端和服务端
- 支持多平台构建（Windows、Linux、macOS）
- 支持多架构（x64、ARM64 for macOS）
- 生成可执行文件并上传为Artifacts
- 在发布Release时自动打包所有平台的文件

**支持的平台：**
- Linux (amd64)
- Windows (amd64)
- macOS (amd64, arm64)

### 2. `.github/workflows/test-build.yml` - 测试构建工作流

这个工作流用于：
- 在推送到 `main` 或 `develop` 分支时触发
- 在创建Pull Request时触发

**功能：**
- 验证代码可以在所有平台上正常编译
- 运行测试用例
- 不生成可执行文件，仅用于验证

## 使用方法

### 自动触发

1. **推送代码**：当你推送代码到 `main` 或 `develop` 分支时，工作流会自动触发
2. **创建PR**：当你创建Pull Request时，会运行测试构建
3. **发布Release**：当你在GitHub上创建Release时，会运行完整构建并生成发布包

### 手动触发

你也可以手动触发工作流：

1. 进入GitHub仓库页面
2. 点击 "Actions" 标签
3. 选择要运行的工作流
4. 点击 "Run workflow" 按钮

## 构建产物

### 客户端文件
- `smart-finder-client-linux-amd64` - Linux版本
- `smart-finder-client-windows-amd64.exe` - Windows版本
- `smart-finder-client-darwin-amd64` - macOS Intel版本
- `smart-finder-client-darwin-arm64` - macOS Apple Silicon版本

### 服务端文件
- `smart-finder-gateway-linux-amd64` - Linux版本
- `smart-finder-gateway-windows-amd64.exe` - Windows版本
- `smart-finder-gateway-darwin-amd64` - macOS Intel版本
- `smart-finder-gateway-darwin-arm64` - macOS Apple Silicon版本

## 环境要求

- Go 1.24
- CGO支持（用于SQLite等依赖）
- 各平台的编译工具链

## 注意事项

1. **CGO依赖**：项目使用了SQLite等需要CGO的依赖，所以构建时需要相应的C编译器
2. **跨平台构建**：使用Go的交叉编译功能，在Linux上构建Windows和macOS版本
3. **文件大小优化**：使用 `-ldflags="-s -w"` 来减小可执行文件大小
4. **缓存优化**：使用GitHub Actions的Go模块缓存来加速构建
5. **构建命令**：使用 `go build .` 而不是 `go build ./main.go` 来确保包含所有Go文件

## 问题解决

### `getPublicFileServer` 未定义错误

如果遇到 `undefined: getPublicFileServer` 错误，这是因为构建命令没有包含所有Go文件。解决方案：

1. **使用正确的构建命令**：
   ```bash
   # 错误的方式
   go build ./main.go
   
   # 正确的方式
   go build .
   ```

2. **确保所有Go文件在同一包中**：`static.go` 和 `main.go` 都应该是 `package main`

3. **检查文件位置**：确保 `static.go` 文件在正确的目录中

## 故障排除

### 常见问题

1. **CGO编译失败**
   - 确保安装了相应的C编译器
   - Linux: `gcc`, `libc6-dev`
   - macOS: `xcode-select --install`
   - Windows: 通常已预装

2. **依赖下载失败**
   - 检查网络连接
   - 验证 `go.mod` 文件是否正确
   - 尝试清理Go模块缓存

3. **跨平台构建失败**
   - 确保Go版本支持目标平台
   - 检查目标平台的CGO依赖是否可用

### 调试方法

1. 查看GitHub Actions日志
2. 在本地使用相同的环境变量进行测试
3. 检查Go版本和依赖版本兼容性 