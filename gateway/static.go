package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed public/*
var publicFS embed.FS

// getPublicFileServer 获取公共文件服务器
func getPublicFileServer() http.Handler {
	// 获取public子目录
	publicDir, err := fs.Sub(publicFS, "public")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(publicDir))
}
