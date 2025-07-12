package types

// FileInfo 文件信息
type FileInfo struct {
	MD5        string `json:"md5"`
	Path       string `json:"path"`
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Version   string `json:"version"`
}

// ClientStatus 客户端状态
type ClientStatus struct {
	Indexing      bool  `json:"indexing"`
	FileCount     int   `json:"fileCount"`
	IndexingTotal int64 `json:"indexingTotal"`
	IndexingDone  int64 `json:"indexingDone"`
}
