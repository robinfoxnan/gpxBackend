package controller

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"hash"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Session struct {
	File     *os.File
	Filename string
	Tempname string // 临时文件名
	UuidName string // 唯一文件名
	hash     hash.Hash
	Lock     sync.Mutex
}

var sessions = make(map[string]*Session)
var sessionLock sync.Mutex

func getSession(sessionID string) *Session {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	return sessions[sessionID]
}

func setSession(sessionID string, session *Session) {
	sessionLock.Lock()
	defer sessionLock.Unlock()
	sessions[sessionID] = session
}

func closeSessionFile(session *Session, rename bool) {
	if session.File != nil {
		session.File.Close()
		session.File = nil

		if rename {
			renameSessionFile(session)
		}
		session.Filename = ""
		session.Tempname = ""
		session.UuidName = ""
	}
}

func renameSessionFile(session *Session) error {
	// 将临时文件重命名为正式文件名
	tempFilePath := filepath.Join(uploadPath, session.Tempname)
	newFilePath := filepath.Join(uploadPath, session.UuidName)

	err := os.Rename(tempFilePath, newFilePath)
	if err != nil {
		fmt.Println(err)
		// 处理错误，可能需要告诉客户端发生了错误
		// 例如：c.JSON(http.StatusInternalServerError, gin.H{"error": "Error renaming file"})
		return err
	}

	fmt.Printf("重命名文件：%s -> %s \n", tempFilePath, newFilePath)
	return nil
}

// 长传需要设置的属性包括：
// sid, chunkNumber, totalChunks, filename, file
func handleUploadChunk1(c *gin.Context) {
	sessionID := c.PostForm("sid")

	// 获取或创建会话
	session := getSession(sessionID)
	if session == nil {
		session = &Session{}
		setSession(sessionID, session)
	}

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SessionID is required"})
		return
	}

	session.Lock.Lock()
	defer session.Lock.Unlock()

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
	// 获取 MD5 值
	md5Code := c.PostForm("md5")
	fmt.Printf("md5= %s \n", md5Code)

	// 获取文件块
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 第一个分片时候确认关闭之前的文件
	if chunkNumber == 0 {
		closeSessionFile(session, false)
	}

	// 其他的片先到了
	if chunkNumber != 0 && session.File == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload chunk 0 first"})
		return
	}

	// 检查文件名是否一致
	if chunkNumber != 0 {
		if session.Filename != filename {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not the same file name in chunks"})
			return
		}
	}

	if session.File == nil {
		// 用临时文件名保存
		session.Filename = filename
		session.UuidName = filename
		session.Tempname = session.UuidName + ".temp"
		session.hash = md5.New()
		fmt.Printf("第一片:%s \n", session.Filename)

		// 这里还需要使用字典法重新确定目录，测试先用同一个目录
		chunkDir := uploadPath

		// 确保长传目录存在
		err = os.MkdirAll(chunkDir, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating chunk directory"})
			return
		}

		// 创建或打开文件
		filePath := filepath.Join(chunkDir, session.Tempname)
		session.File, err = os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating file"})
			return
		}
	}

	// 保存分块文件
	err = saveChunkInSame(session.File, &file)
	if err != nil {
		closeSessionFile(session, false)
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chunk file"})
		return // 返回错误
	} else {
		file.Seek(0, 0)
		if _, err := io.Copy(session.hash, file); err != nil {
			fmt.Println("添加哈希分片过程错误")
		}
	}

	// 检查是否所有分块都已上传
	if chunkNumber == totalChunks-1 {
		closeSessionFile(session, true)

		// 计算最终的 MD5 散列值
		hashInBytes := session.hash.Sum(nil)
		hashString := hex.EncodeToString(hashInBytes)
		fmt.Printf("接收的md5 = %s \n", hashString)

		//str, _ := calculateFileMD5(filepath.Join(uploadPath, filename))
		//fmt.Printf("接收的md5 = %s \n", str)

		c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
	}
}

func saveChunkInSame(tempFile *os.File, file *multipart.File) error {

	if tempFile == nil || file == nil {
		return errors.New("tempFile or file is nil")
	}
	// 设置文件指针的位置
	//newOffset, err := file.Seek(offset, 0) // 从文件开头开始计算偏移量
	//if err != nil {
	//	fmt.Println("Error seeking file:", err)
	//	return
	//}

	_, err := io.Copy(tempFile, *file)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func calculateFileMD5(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}
	sz := int64(1) << 20
	if fileInfo.Size() < sz {
		return calculateFileMD5Small(filePath)
	} else {
		return calculateFileMD5Chunk(filePath)
	}
}

// 一次性计算file的md5
func calculateFileMD5Small(filePath string) (string, error) {
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

// 一次性计算file的md5
func calculateFileMD5Chunk(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()

	// 逐个添加 chunk 并计算散列值
	const chunkSize = 8192
	buffer := make([]byte, chunkSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		hash.Write(buffer[:n])
	}

	// 计算最终的 MD5 散列值
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
}
