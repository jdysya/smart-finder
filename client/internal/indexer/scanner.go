package indexer

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"smart-finder/client/internal/core"
)

var (
	Indexing      bool
	IndexingTotal int
	IndexingDone  int
)

func Scanner(db *sql.DB, root string) error {
	Indexing = true
	defer func() { Indexing = false }()
	// 先统计总文件数
	total := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		total++
		return nil
	})
	IndexingTotal = total
	IndexingDone = 0
	// 正式索引
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("扫描文件/目录时出错: %s, 错误: %v, 已跳过", path, err)
			return nil
		}
		if info.IsDir() {
			return nil // 跳过目录本身
		}
		md5sum, err := CalculateMD5(path)
		if err != nil {
			log.Printf("计算 MD5 时出错: %s, 错误: %v, 已跳过", path, err)
			return nil
		}
		fileIndex := core.FileIndex{
			MD5:        md5sum,
			Path:       path,
			Filename:   info.Name(),
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		}
		// 插入或更新
		_, err = db.Exec(`
            INSERT INTO files (md5, path, filename, size, modified_at)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(md5) DO UPDATE SET
                path=excluded.path,
                filename=excluded.filename,
                size=excluded.size,
                modified_at=excluded.modified_at
        `, fileIndex.MD5, fileIndex.Path, fileIndex.Filename, fileIndex.Size, fileIndex.ModifiedAt)
		if err != nil {
			log.Printf("索引文件失败: %s, 错误: %v", path, err)
		}
		IndexingDone++
		return nil
	})
}
