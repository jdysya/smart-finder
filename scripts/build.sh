#!/bin/bash

# Smart Finder 构建脚本
# 用于本地测试构建过程

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go版本
check_go_version() {
    print_info "检查Go版本..."
    go version
    if [ $? -ne 0 ]; then
        print_error "Go未安装或不在PATH中"
        exit 1
    fi
}

# 构建客户端
build_client() {
    print_info "构建客户端..."
    cd client
    
    # 清理之前的构建
    rm -f smart-finder-client-*
    
    # 构建当前平台
    go build -ldflags="-s -w" -o "smart-finder-client-$(go env GOOS)-$(go env GOARCH)" .
    
    if [ $? -eq 0 ]; then
        print_info "客户端构建成功: smart-finder-client-$(go env GOOS)-$(go env GOARCH)"
    else
        print_error "客户端构建失败"
        exit 1
    fi
    
    cd ..
}

# 构建服务端
build_gateway() {
    print_info "构建服务端..."
    cd gateway
    
    # 清理之前的构建
    rm -f smart-finder-gateway-*
    
    # 构建当前平台
    go build -ldflags="-s -w" -o "smart-finder-gateway-$(go env GOOS)-$(go env GOARCH)" .
    
    if [ $? -eq 0 ]; then
        print_info "服务端构建成功: smart-finder-gateway-$(go env GOOS)-$(go env GOARCH)"
    else
        print_error "服务端构建失败"
        exit 1
    fi
    
    cd ..
}

# 交叉编译测试
cross_compile_test() {
    print_info "测试交叉编译..."
    
    # 测试Linux构建
    print_info "测试Linux amd64构建..."
    cd gateway
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "smart-finder-gateway-linux-amd64" .
    if [ $? -eq 0 ]; then
        print_info "Linux构建成功"
        rm -f smart-finder-gateway-linux-amd64
    else
        print_error "Linux构建失败"
        exit 1
    fi
    cd ..
    
    # 测试Windows构建
    print_info "测试Windows amd64构建..."
    cd gateway
    GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o "smart-finder-gateway-windows-amd64.exe" .
    if [ $? -eq 0 ]; then
        print_info "Windows构建成功"
        rm -f smart-finder-gateway-windows-amd64.exe
    else
        print_error "Windows构建失败"
        exit 1
    fi
    cd ..
    
    # 测试macOS构建
    print_info "测试macOS amd64构建..."
    cd gateway
    GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o "smart-finder-gateway-darwin-amd64" .
    if [ $? -eq 0 ]; then
        print_info "macOS amd64构建成功"
        rm -f smart-finder-gateway-darwin-amd64
    else
        print_error "macOS amd64构建失败"
        exit 1
    fi
    cd ..
    
    # 测试macOS ARM64构建
    print_info "测试macOS arm64构建..."
    cd gateway
    GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o "smart-finder-gateway-darwin-arm64" .
    if [ $? -eq 0 ]; then
        print_info "macOS arm64构建成功"
        rm -f smart-finder-gateway-darwin-arm64
    else
        print_error "macOS arm64构建失败"
        exit 1
    fi
    cd ..
}

# 运行测试
run_tests() {
    print_info "运行测试..."
    # 在Go工作区中运行测试
    go test ./client/...
    go test ./gateway/...
    go test ./shared/...
    if [ $? -eq 0 ]; then
        print_info "所有测试通过"
    else
        print_error "测试失败"
        exit 1
    fi
}

# 主函数
main() {
    print_info "开始Smart Finder构建测试..."
    
    check_go_version
    run_tests
    build_client
    build_gateway
    cross_compile_test
    
    print_info "所有构建测试完成！"
}

# 脚本入口
if [ "$1" = "client" ]; then
    check_go_version
    build_client
elif [ "$1" = "gateway" ]; then
    check_go_version
    build_gateway
elif [ "$1" = "cross" ]; then
    check_go_version
    cross_compile_test
elif [ "$1" = "test" ]; then
    check_go_version
    run_tests
else
    main
fi 