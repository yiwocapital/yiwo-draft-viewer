package file

import (
	"fmt"
	"os"
)

const MaxFileSize = 5 * 1024 * 1024 // 5MB

func Load(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("cannot stat file: %w", err)
	}
	if info.Size() > MaxFileSize {
		return "", fmt.Errorf("file too large (%d bytes, max %d)", info.Size(), MaxFileSize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %w", err)
	}
	return string(data), nil
}
