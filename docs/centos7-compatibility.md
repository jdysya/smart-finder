# CentOS 7 兼容性说明

## 问题描述

之前的构建版本在CentOS 7上无法运行，主要原因是：

1. **CGO依赖**: 启用了CGO (`CGO_ENABLED=1`)，导致二进制文件依赖动态链接库
2. **动态链接**: 没有使用静态链接，依赖系统库版本
3. **架构兼容性**: 没有指定合适的AMD64架构版本

## 解决方案

### 1. 静态链接构建

修改构建配置，使用静态链接：

```yaml
env:
  CGO_ENABLED: 0  # 禁用CGO
```

构建命令：
```bash
go build -ldflags="-s -w -extldflags=-static" -tags="osusergo,netgo" -o output .
```

### 2. 架构兼容性

设置 `GOAMD64=v1` 确保与旧版本CPU兼容：

```bash
export GOAMD64=v1
```

### 3. 构建标签

使用以下构建标签确保纯Go实现：
- `osusergo`: 使用Go实现的用户/组查找
- `netgo`: 使用Go实现的网络库

## 构建脚本

使用 `scripts/build-centos7.sh` 脚本进行本地构建：

```bash
./scripts/build-centos7.sh
```

## 验证方法

### 1. 检查文件类型
```bash
file smart-finder-client-linux-amd64-centos7
```

应该显示：`ELF 64-bit LSB executable, statically linked`

### 2. 检查动态依赖
```bash
ldd smart-finder-client-linux-amd64-centos7
```

应该显示：`not a dynamic executable` 或 `静态链接 - 无动态依赖`

### 3. 在CentOS 7上测试
```bash
# 在CentOS 7系统上
chmod +x smart-finder-client-linux-amd64-centos7
./smart-finder-client-linux-amd64-centos7 --help
```

## 支持的版本

修改后的构建支持以下系统：

- **CentOS 7+** / **RHEL 7+**
- **Ubuntu 16.04+**
- **Debian 9+**
- **Windows 7+**
- **macOS 10.12+**

## 注意事项

1. **文件大小**: 静态链接会增加二进制文件大小
2. **性能**: 静态链接通常有轻微的性能提升
3. **兼容性**: 确保与旧版本系统的兼容性
4. **SQLite驱动**: 使用 `modernc.org/sqlite` 纯Go实现，无需CGO支持

## SQLite 驱动变更

项目已从 `github.com/mattn/go-sqlite3` 迁移到 `modernc.org/sqlite`：

### 变更原因
- **CGO依赖**: 原驱动需要CGO支持，无法实现静态链接
- **跨平台兼容**: 新驱动为纯Go实现，支持所有平台
- **CentOS 7兼容**: 避免动态库依赖问题

### 技术细节
- **驱动名称**: 从 `sqlite3` 改为 `sqlite`
- **导入路径**: 从 `github.com/mattn/go-sqlite3` 改为 `modernc.org/sqlite`
- **功能兼容**: 完全兼容原有SQLite功能

## 故障排除

如果仍然遇到问题：

1. **权限问题**: 确保文件有执行权限
   ```bash
   chmod +x smart-finder-*
   ```

2. **SELinux**: 在CentOS/RHEL上可能需要调整SELinux策略
   ```bash
   setsebool -P httpd_can_network_connect 1
   ```

3. **防火墙**: 确保端口没有被防火墙阻止
   ```bash
   firewall-cmd --add-port=8080/tcp --permanent
   firewall-cmd --reload
   ``` 