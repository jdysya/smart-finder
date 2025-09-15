package indexer

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"smart-finder/client/internal/core"
	"smart-finder/client/internal/db"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ScanStatus 扫描状态
type ScanStatus struct {
	IsScanning    bool      `json:"is_scanning"`
	StartTime     time.Time `json:"start_time"`
	TotalFiles    int64     `json:"total_files"`
	ProcessedFiles int64    `json:"processed_files"`
	SkippedFiles  int64     `json:"skipped_files"`
	ErrorFiles    int64     `json:"error_files"`
	DeletedFiles  int64     `json:"deleted_files"`
	CurrentDir    string    `json:"current_dir"`
	Progress      float64   `json:"progress"`
	ElapsedTime   string    `json:"elapsed_time"`
}

// FileRecord 文件记录结构
type FileRecord struct {
	Path       string
	MD5        string
	Size       int64
	ModifiedAt time.Time
}

// ScheduledScanner 定时扫描器
type ScheduledScanner struct {
	dbConn          *sql.DB
	scanInterval    time.Duration
	batchSize       int
	maxConcurrency  int
	status          ScanStatus
	statusMu        sync.RWMutex
	stopChan        chan struct{}
	manualTrigger   chan struct{}
	isRunning       int32
	dbMutex         sync.Mutex  // 添加数据库操作互斥锁
}

// NewScheduledScanner 创建新的定时扫描器
func NewScheduledScanner(dbConn *sql.DB, interval time.Duration) *ScheduledScanner {
	return &ScheduledScanner{
		dbConn:         dbConn,
		scanInterval:   interval,
		batchSize:      500,  // 批量处理大小
		maxConcurrency: 3,    // 最大并发数，避免过度占用资源
		stopChan:       make(chan struct{}),
		manualTrigger:  make(chan struct{}, 1),
	}
}

// Start 启动定时扫描器
func (s *ScheduledScanner) Start() {
	if !atomic.CompareAndSwapInt32(&s.isRunning, 0, 1) {
		log.Println("扫描器已在运行中")
		return
	}

	log.Println("启动定时扫描器...")
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	// 启动时立即执行一次扫描
	go s.performScan()

	for {
		select {
		case <-ticker.C:
			go s.performScan()
		case <-s.manualTrigger:
			go s.performScan()
		case <-s.stopChan:
			atomic.StoreInt32(&s.isRunning, 0)
			log.Println("定时扫描器已停止")
			return
		}
	}
}

// Stop 停止扫描器
func (s *ScheduledScanner) Stop() {
	close(s.stopChan)
}

// TriggerManualScan 手动触发扫描
func (s *ScheduledScanner) TriggerManualScan() {
	select {
	case s.manualTrigger <- struct{}{}:
		log.Println("手动触发扫描")
	default:
		log.Println("扫描已在进行中，忽略手动触发")
	}
}

// GetStatus 获取当前扫描状态
func (s *ScheduledScanner) GetStatus() ScanStatus {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	
	status := s.status
	if status.IsScanning && !status.StartTime.IsZero() {
		status.ElapsedTime = time.Since(status.StartTime).Round(time.Second).String()
		if status.TotalFiles > 0 {
			status.Progress = float64(status.ProcessedFiles+status.SkippedFiles) / float64(status.TotalFiles) * 100
		}
	}
	return status
}

// updateStatus 更新扫描状态
func (s *ScheduledScanner) updateStatus(update func(*ScanStatus)) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	update(&s.status)
}

// performScan 执行扫描
func (s *ScheduledScanner) performScan() {
	// 检查是否已在扫描中
	s.statusMu.RLock()
	isScanning := s.status.IsScanning
	s.statusMu.RUnlock()
	
	if isScanning {
		log.Println("扫描已在进行中，跳过本次扫描")
		return
	}

	log.Println("开始定时扫描...")
	startTime := time.Now()

	// 初始化扫描状态
	s.updateStatus(func(status *ScanStatus) {
		*status = ScanStatus{
			IsScanning:  true,
			StartTime:   startTime,
			CurrentDir:  "准备中...",
		}
	})

	defer func() {
		s.updateStatus(func(status *ScanStatus) {
			status.IsScanning = false
			status.CurrentDir = "扫描完成"
		})
	}()

	// 获取监控目录
	monitoredDirs, err := db.GetMonitoredDirectories(s.dbConn)
	if err != nil {
		log.Printf("获取监控目录失败: %v", err)
		return
	}

	if len(monitoredDirs) == 0 {
		log.Println("没有配置监控目录，跳过扫描")
		return
	}

	// 获取忽略模式
	ignorePatterns, err := db.GetIgnoredPatterns(s.dbConn)
	if err != nil {
		log.Printf("获取忽略模式失败: %v", err)
		// 继续执行，不中断扫描
	}

	// 标记扫描开始
	if err := s.markScanStart(); err != nil {
		log.Printf("标记扫描开始失败: %v", err)
		return
	}

	// 第一阶段：统计总文件数
	s.updateStatus(func(status *ScanStatus) {
		status.CurrentDir = "统计文件数量..."
	})
	
	totalFiles := s.countTotalFiles(monitoredDirs, ignorePatterns)
	s.updateStatus(func(status *ScanStatus) {
		status.TotalFiles = totalFiles
	})

	// 第二阶段：扫描文件
	for _, dir := range monitoredDirs {
		s.updateStatus(func(status *ScanStatus) {
			status.CurrentDir = dir
		})
		s.scanDirectory(dir, ignorePatterns)
	}

	// 第三阶段：清理不存在的文件
	s.updateStatus(func(status *ScanStatus) {
		status.CurrentDir = "清理过时文件..."
	})
	
	deletedCount, err := s.cleanupMissingFiles()
	if err != nil {
		log.Printf("清理过时文件失败: %v", err)
	} else {
		s.updateStatus(func(status *ScanStatus) {
			status.DeletedFiles = deletedCount
		})
	}

	duration := time.Since(startTime)
	finalStatus := s.GetStatus()
	log.Printf("扫描完成 - 总计: %d, 处理: %d, 跳过: %d, 错误: %d, 删除: %d, 耗时: %v",
		finalStatus.TotalFiles, finalStatus.ProcessedFiles, finalStatus.SkippedFiles, 
		finalStatus.ErrorFiles, finalStatus.DeletedFiles, duration)
}

// countTotalFiles 统计总文件数
func (s *ScheduledScanner) countTotalFiles(monitoredDirs []string, ignorePatterns []string) int64 {
	var total int64
	
	for _, rootDir := range monitoredDirs {
		filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			if s.shouldIgnore(path, info, ignorePatterns) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			
			if !info.IsDir() {
				atomic.AddInt64(&total, 1)
			}
			return nil
		})
	}
	
	return total
}

// scanDirectory 扫描目录
func (s *ScheduledScanner) scanDirectory(rootDir string, ignorePatterns []string) {
	var fileBatch []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			atomic.AddInt64(&s.status.ErrorFiles, 1)
			return nil
		}

		if s.shouldIgnore(path, info, ignorePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		fileBatch = append(fileBatch, path)

		// 批量处理
		if len(fileBatch) >= s.batchSize {
			s.processBatch(fileBatch)
			fileBatch = fileBatch[:0]
		}

		return nil
	})

	if err != nil {
		log.Printf("扫描目录 %s 失败: %v", rootDir, err)
	}

	// 处理剩余文件
	if len(fileBatch) > 0 {
		s.processBatch(fileBatch)
	}
}

// processBatch 批量处理文件
func (s *ScheduledScanner) processBatch(filePaths []string) {
	// 批量获取现有文件信息
	existingFiles, err := s.getExistingFileInfo(filePaths)
	if err != nil {
		log.Printf("获取现有文件信息失败: %v", err)
	}

	// 减少并发数以降低数据库锁竞争
	concurrency := min(s.maxConcurrency, 2) // 限制最大并发数为2
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, filePath := range filePaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			s.processFile(path, existingFiles)
		}(filePath)
	}

	wg.Wait()
}

// processFile 处理单个文件
func (s *ScheduledScanner) processFile(filePath string, existingFiles map[string]FileRecord) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("获取文件信息失败 %s: %v", filePath, err)
		atomic.AddInt64(&s.status.ErrorFiles, 1)
		return
	}

	// 标记文件存在
	if err := s.markFileExists(filePath); err != nil {
		log.Printf("标记文件存在失败 %s: %v", filePath, err)
	}

	// 检查是否需要重新计算MD5
	if existing, found := existingFiles[filePath]; found {
		if existing.Size == fileInfo.Size() && existing.ModifiedAt.Equal(fileInfo.ModTime()) {
			// 文件未变化，跳过
			atomic.AddInt64(&s.status.SkippedFiles, 1)
			return
		}
	}

	// 计算MD5
	md5sum, err := CalculateMD5(filePath)
	if err != nil {
		log.Printf("计算MD5失败 %s: %v", filePath, err)
		atomic.AddInt64(&s.status.ErrorFiles, 1)
		return
	}

	// 更新数据库
	fileIndex := core.FileIndex{
		MD5:        md5sum,
		Path:       filePath,
		Filename:   fileInfo.Name(),
		Size:       fileInfo.Size(),
		ModifiedAt: fileInfo.ModTime(),
	}

	s.dbMutex.Lock()
	_, err = s.dbConn.Exec(`
		INSERT OR REPLACE INTO files (md5, path, filename, size, modified_at, scan_flag)
		VALUES (?, ?, ?, ?, ?, 1)
	`, fileIndex.MD5, fileIndex.Path, fileIndex.Filename, fileIndex.Size, fileIndex.ModifiedAt)
	s.dbMutex.Unlock()

	if err != nil {
		log.Printf("更新数据库失败 %s: %v", filePath, err)
		atomic.AddInt64(&s.status.ErrorFiles, 1)
		return
	}

	atomic.AddInt64(&s.status.ProcessedFiles, 1)
}

// shouldIgnore 检查是否应该忽略文件/目录
func (s *ScheduledScanner) shouldIgnore(path string, info os.FileInfo, patterns []string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, info.Name())
		if matched {
			return true
		}
		// 也检查完整路径
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}

// markScanStart 标记扫描开始
func (s *ScheduledScanner) markScanStart() error {
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	_, err := s.dbConn.Exec("UPDATE files SET scan_flag = 0")
	return err
}

// markFileExists 标记文件存在
func (s *ScheduledScanner) markFileExists(filePath string) error {
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	_, err := s.dbConn.Exec("UPDATE files SET scan_flag = 1 WHERE path = ?", filePath)
	return err
}

// cleanupMissingFiles 清理不存在的文件
func (s *ScheduledScanner) cleanupMissingFiles() (int64, error) {
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	result, err := s.dbConn.Exec("DELETE FROM files WHERE scan_flag = 0")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// getExistingFileInfo 批量获取现有文件信息
func (s *ScheduledScanner) getExistingFileInfo(paths []string) (map[string]FileRecord, error) {
	if len(paths) == 0 {
		return make(map[string]FileRecord), nil
	}

	// 构建批量查询
	placeholders := make([]string, len(paths))
	args := make([]interface{}, len(paths))
	for i, path := range paths {
		placeholders[i] = "?"
		args[i] = path
	}

	query := fmt.Sprintf("SELECT path, md5, size, modified_at FROM files WHERE path IN (%s)",
		strings.Join(placeholders, ","))

	rows, err := s.dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]FileRecord)
	for rows.Next() {
		var record FileRecord
		err := rows.Scan(&record.Path, &record.MD5, &record.Size, &record.ModifiedAt)
		if err != nil {
			continue
		}
		result[record.Path] = record
	}

	return result, nil
}

// 全局扫描器实例
var (
	GlobalScheduler *ScheduledScanner
	schedulerMu     sync.Mutex
)

// InitGlobalScheduler 初始化全局扫描器
func InitGlobalScheduler(dbConn *sql.DB, interval time.Duration) {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	
	if GlobalScheduler != nil {
		GlobalScheduler.Stop()
	}
	
	GlobalScheduler = NewScheduledScanner(dbConn, interval)
}

// GetGlobalScheduler 获取全局扫描器
func GetGlobalScheduler() *ScheduledScanner {
	schedulerMu.Lock()
	defer schedulerMu.Unlock()
	return GlobalScheduler
}