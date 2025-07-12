package core

import "time"

type FileIndex struct {
	MD5        string    `json:"md5" gorm:"primaryKey"`
	Path       string    `json:"path"`
	Filename   string    `json:"filename"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}
