package file

import (
	"os"
)

type FileHandler struct{}

func (fh *FileHandler) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (fh *FileHandler) WriteFile(filePath string, content string) error {
	return os.WriteFile(filePath, []byte(content), os.ModePerm)
}
