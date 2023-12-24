package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"zhituBackend/common"
)

func InitHanlder(router *gin.Engine) {

	// 静态页面映射
	router.StaticFile("/favicon.ico", "./web/static/favicon.ico")
	router.Static("/static", "./web/static")

	// 遍历目录，生成名字模板
	loadTemplateDir("./web", router)

	//router.GET("/", indexHandler)
	router.GET("/test", indexHandler)
	router.GET("/ver", versionHandler)

	router.GET("/download", downloadPageHandler) // 从目录中获取下载文件列表
	router.GET("/download/:filename", fileDownloadHandler)

	router.GET("/file/:filename", fileDownloadExHandler)
	//
	//
	router.POST("/uploadfile", HandleUpload)
	router.POST("/uploadchunk", handleUploadChunk1)

	// router.GET("/ws", ws.InitWebSocket)

	initUserHandlerV1(router)
	initGpxHandlerV1(router)
	InitNewsHandler(router)

}

// WrapHttpHandler 将http.HandlerFunc类型的函数包装为gin.HandlerFunc类型
func WrapHttpHandler(fn http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将gin.Context转换为http.ResponseWriter和*http.Request
		fn(c.Writer, c.Request)
	}
}

// "./templates"
// router.LoadHTMLGlob("web/**/*")
func loadTemplateDir(filename string, router *gin.Engine) {
	templateFiles := make([]string, 0, 32)
	filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			//fmt.Println(path)
			templateFiles = append(templateFiles, path)
		}
		return nil
	})

	fmt.Println(templateFiles)
	router.LoadHTMLFiles(templateFiles...)
}

func indexHandler(c *gin.Context) {
	c.HTML(200, "about.html", nil)
}

func favicon(c *gin.Context) {
	c.Redirect(202, "static/favicon.ico")
}

func versionHandler(c *gin.Context) {
	protocol := c.Request.Proto
	fmt.Printf("Request protocol: %s\n", protocol)
	str := fmt.Sprintf("协议 %s", protocol)
	c.JSON(http.StatusOK, gin.H{"message": str})
}

// downloadHandler 处理/download路由，生成下载页面
func downloadPageHandler(c *gin.Context) {

	// 获取download目录下的所有文件
	files, err := ioutil.ReadDir("web/download")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error reading download directory: %s", err))
		return
	}

	// 生成下载页面的HTML内容
	var downloadLinks strings.Builder
	for _, file := range files {
		if file.IsDir() {
			// 如果是目录，跳过
			continue
		}
		fileName := file.Name()
		downloadLinks.WriteString(fmt.Sprintf("<p><a href=\"/download/%s\" download=\"%s\">%s</a></p>", fileName, fileName, fileName))
	}

	// 渲染下载页面
	//c.HTML(http.StatusOK, "download.html", gin.H{"downloadLinks": downloadLinks.String()})
	c.HTML(http.StatusOK, "download.html", gin.H{"downloadLinks": template.HTML(downloadLinks.String())})

}

// fileDownloadHandler 处理文件下载
// http://localhost/download/favicon.ico?sid=1111
func fileDownloadHandler(c *gin.Context) {

	//设置默认值
	//user := c.DefaultQuery("username", "test")
	sid := c.Query("sid")
	fmt.Println("sid = ", sid)

	filename := c.Param("filename")
	fmt.Println(filename)
	filePath := filepath.Join("./web/download", filename)

	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// 提供文件下载
	c.File(filePath)
}

// 扩展下载
func fileDownloadExHandler(c *gin.Context) {

	//设置默认值
	//user := c.DefaultQuery("username", "test")
	sid := c.Query("sid")
	fmt.Println("sid = ", sid)

	filename := c.Param("filename")
	fmt.Println(filename)

	//filePath := filepath.Join("./web/download", filename)

	filePath, err := common.FileName2FilePath("", filename, false)
	if err != nil {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// 检查文件是否存在
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	// 提供文件下载
	c.File(filePath)
}

// 优先从hueader取参数，之后是从参数，最次从匹配的路径
func queryParamCommon(c *gin.Context, key string) string {

	v := c.Query(key)
	if len(v) != 0 {
		return v
	}

	v = c.GetHeader(key)
	if len(v) != 0 {
		return v
	}

	v = c.Param(key)
	if len(v) != 0 {
		return v
	}

	return v
}
