package indexer

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	buf := make([]byte, 4*1024*1024) // 4MB分块
	for {
		n, err := file.Read(buf)
		if n > 0 {
			if _, wErr := hash.Write(buf[:n]); wErr != nil {
				return "", wErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
