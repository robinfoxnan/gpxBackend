package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"zhituBackend/common"
)

const maxUploadSize = 10 * 1024 * 1024 // 10 MB
const uploadPath = "./web/filestore"

//	func main() {
//		router := gin.Default()
//
//		router.POST("/upload", handleUpload)
//
//		router.Run(":8080")
//	}

func HandleUpload(c *gin.Context) {

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "fail",
			"des":   err.Error(),
		})
		return
	}
	defer file.Close()

	// 提取隐藏字段，如果是网页来的返回网页，如果不是，则返回JSON
	srcValue := c.PostForm("src")
	common.Logger.Info("src value :", zap.String("from ", srcValue))

	fileSize, _ := strconv.Atoi(c.Request.Header.Get("Content-Length"))
	if int64(fileSize) > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "fail",
			"des":   fmt.Sprintf("File size exceeds the limit of %d MB", maxUploadSize/(1024*1024)),
		})
		return
	}

	// Create a unique filename
	filename := header.Filename
	//savePath := filepath.Join(uploadPath, filename)

	// Save the file to disk
	newFileName, err := saveFile(file, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"state": "fail",
			"des":   "Error saving file"})
		return
	}

	if srcValue == "webpage" {
		c.HTML(http.StatusOK, "upload_success.html", gin.H{
			"des":      "File uploaded successfully",
			"filename": filename,
			"newname":  newFileName,
			"state":    "OK",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"des":      "File uploaded successfully",
			"filename": filename,
			"newname":  newFileName,
			"state":    "OK",
		})
	}

}

func saveFile(file multipart.File, filename string) (string, error) {

	_, ext := common.DepartFileName(filename)
	//fmt.Println("recv upload: ", filename)
	common.Logger.Info("recv upload: ", zap.String("filename", filename))
	mainNameNew := common.NextFileName() // 雪花算法
	fileFullName, err := common.FileNameExt2FilePath("", mainNameNew, ext, true)
	if err != nil {
		return fileFullName, err
	}
	//err := os.MkdirAll(uploadPath, os.ModePerm)
	//if err != nil {
	//	return err
	//}

	dst, err := os.Create(fileFullName)
	if err != nil {
		return fileFullName, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		fmt.Println(err)
		return fileFullName, err
	}

	return mainNameNew + ext, nil
}

func testFilename() {
	common.InitSnow(1)
	mainName := common.NextFileName()
	fmt.Println(mainName)
	fileName, err := common.FileNameExt2FilePath("", mainName, ".jpg", false)
	fmt.Println(fileName, err)

}
