package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func handleUploadChunk(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 获取块的相关信息
	chunkNumber, err := strconv.Atoi(c.PostForm("chunkNumber"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunk number"})
		fmt.Println(err)
		return
	}

	totalChunks, err := strconv.Atoi(c.PostForm("totalChunks"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid total chunks"})
		fmt.Println(err)
		return
	}

	// 获取文件名
	filename := c.PostForm("filename")

	// 创建保存分块的目录
	chunkDir := filepath.Join(uploadPath, "chunks_"+filename)

	err = os.MkdirAll(chunkDir, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating chunk directory"})
		return
	}

	// 保存分块文件
	saveChunk(chunkDir, chunkNumber, c, &file)

	// 检查是否所有分块都已上传
	if chunkNumber == totalChunks-1 {
		// 所有分块已上传，将它们组合为完整文件
		err = combineChunks(chunkDir, filename)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error combining chunks"})
			return
		}

		// 等待一段时间以确保文件被释放
		time.Sleep(2 * time.Second)
		// 清理分块文件夹
		err = os.RemoveAll(chunkDir)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cleaning up chunk directory"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	} else {
		// 通知前端可以继续上传下一个块
		c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully"})
	}
}

func saveChunk(chunkDir string, chunkNumber int, c *gin.Context, file *multipart.File) {
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkNumber))
	dst, err := os.Create(chunkPath)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating chunk file"})
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, *file)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chunk file"})
		return
	}
}

func combineChunks(chunkDir, filename string) error {
	// 创建完整文件
	fullFilePath := filepath.Join(uploadPath, filename)
	fullFile, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	defer fullFile.Close()

	// 遍历分块文件，逐个追加到完整文件
	chunkFiles, err := filepath.Glob(filepath.Join(chunkDir, "chunk_*"))
	if err != nil {
		return err
	}
	for _, chunkFile := range chunkFiles {
		chunk, err := os.Open(chunkFile)
		if err != nil {
			return err
		}
		defer chunk.Close()

		_, err = io.Copy(fullFile, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}
