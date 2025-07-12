package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// CalculateMD5 计算文件的MD5哈希值
func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidateMD5 验证MD5哈希值格式
func ValidateMD5(hash string) bool {
	if len(hash) != 32 {
		return false
	}
	
	for _, char := range hash {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}
