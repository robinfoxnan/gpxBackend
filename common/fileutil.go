package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// 雪花算法一般不会出现小于4字节
func FileNameExt2FilePath(base, mainName, extName string, bCreate bool) (string, error) {

	if len(mainName) < 4 {
		return filepath.Join(base, mainName, extName), nil
	}

	if base == "" {
		// 获取当前工作目录
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			return "", errors.New("")
		}
		base = filepath.Join(currentDir, "web/filestore")
	}

	// 获取文件名的前两个字节
	firstTwoBytes := mainName[:2]
	nextTwoBytes := mainName[2:4]

	newPath := filepath.Join(base, firstTwoBytes, nextTwoBytes)
	// 创建目录, 下载时候不需要只需要检查是否存在
	if bCreate {
		err := os.MkdirAll(newPath, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directories:", err)
			return "", err
		}
	}

	return filepath.Join(newPath, mainName+extName), nil
}

func FileName2FilePath(base, fileName string, bCreate bool) (string, error) {

	baseName := filepath.Base(fileName)
	mainName := baseName[:len(baseName)-len(filepath.Ext(baseName))]
	ext := filepath.Ext(baseName)

	return FileNameExt2FilePath(base, mainName, ext, bCreate)
}

func DepartFileName(fileName string) (string, string) {
	baseName := filepath.Base(fileName)
	mainName := baseName[:len(baseName)-len(filepath.Ext(baseName))]
	ext := filepath.Ext(baseName)
	return mainName, ext
}
