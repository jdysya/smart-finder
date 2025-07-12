package templates

import (
	"embed"
	"html/template"
	"io"
)

//go:embed *.html
var templateFS embed.FS

// TemplateData 模板数据结构
type TemplateData struct {
	Hash         string
	ServerDomain string
}

// Templates 模板管理器
type Templates struct {
	md5Template *template.Template
}

// New 创建新的模板管理器
func New() (*Templates, error) {
	// 解析MD5模板
	md5Template, err := template.ParseFS(templateFS, "md5.html")
	if err != nil {
		return nil, err
	}

	return &Templates{
		md5Template: md5Template,
	}, nil
}

// RenderMD5Page 渲染MD5页面
func (t *Templates) RenderMD5Page(w io.Writer, data TemplateData) error {
	return t.md5Template.Execute(w, data)
}

// GetTemplateFS 获取模板文件系统
func GetTemplateFS() embed.FS {
	return templateFS
}
