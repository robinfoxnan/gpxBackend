package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"zhituBackend/common"
	"zhituBackend/db"
	"zhituBackend/model"
)

func InitNewsHandler(r *gin.Engine) {

	newsGroup := r.Group("/v1/news")
	newsGroup.Use(AuthMiddleware())
	{
		newsGroup.POST("/publish", publishNewsHandler)
		newsGroup.GET("/delete", deleteNewsHandler)

		newsGroup.GET("/recent", findRecentNewsHandler)
		newsGroup.GET("/bytag", findNewsByTagHandler)
		newsGroup.GET("/byloc", findNewsByLocationHandler)

		newsGroup.GET("/setfav", addNewsFavCount)
		newsGroup.GET("/setlike", addNewsLikeCount)
		newsGroup.GET("/sethate", addNewsHate)

		newsGroup.POST("/addcomment", addNewsCommentHandler)
		newsGroup.GET("/deletecomment", deleteNewsCommentHandler)
		newsGroup.GET("/findcomment", findNewsCommentHandler)

		//r.GET("/v1/news/setcommentlike", setCommentLikeHandler)

		newsGroup.POST("/report", setReportHandler) // 投诉，
	}

}

func printErrResult(c *gin.Context, ret string, code int, des string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"state": ret,
		"code":  code,
		"des":   des,
	})
}

func printOkResult(c *gin.Context, code int, des string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"state": "ok",
		"code":  code,
		"des":   des,
	})
}

func publishNewsHandler(c *gin.Context) {
	// 从请求 Body 中提取 JSON 数据
	body, err := c.GetRawData()
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, model.ErrBadParam.Error())
		return
	}

	// 调用 NewsFromJson 函数转为结构体
	news, err := model.NewsFromJsonBytes(body)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, model.ErrBadParam.Error())
		return
	}

	uid := queryParamCommon(c, "uid")
	if news.Uid != uid {
		printErrResult(c, "fail", model.ErrBadParamCode, "uid not consist with uid in news json.")
		return
	}

	// 在这里可以使用得到的 news 结构体进行后续处理
	err = db.MongoClient.SaveNews(news)
	if err != nil {
		common.Logger.Error("save new to mongo err", zap.String("body", string(body)))
	}

	// 应答 200
	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "save news ok",
		"nid":   news.Nid,
	})
}

// 删除一个文档
func deleteNewsHandler(c *gin.Context) {
	nid := queryParamCommon(c, "nid") // 获取查询参数中的tag值
	uid := queryParamCommon(c, "uid")

	// 在这里可以使用得到的 news 结构体进行后续处理
	err := db.MongoClient.DeleteOneNewsByMark(nid, uid)
	if err != nil {
		common.Logger.Error("delete mongo err", zap.String("nid", nid))
	}

	// 应答 200
	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "delete news ok",
		"nid":   nid,
	})
}

func findRecentNewsHandler(c *gin.Context) {

	data, err := db.MongoClient.FindLatestNews()
	if err != nil {
		printErrResult(c, "fail", model.ErrNoDataCode, model.ErrNoData.Error())
		return
	}

	if len(data) < 1 {
		printErrResult(c, "fail", model.ErrNoDataCode, model.ErrNoData.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "find news ok",
		"data":  data,
	})
}

func findNewsByTagHandler(c *gin.Context) {
	tag := c.Query("tag") // 获取查询参数中的tag值
	if tag == "" {
		printErrResult(c, "fail", model.ErrNoDataCode, model.ErrNoData.Error())
		return
	}

	data, err := db.MongoClient.FindNewsByTag(tag)
	if err != nil {
		printErrResult(c, "fail", model.ErrNoDataCode, model.ErrNoData.Error())
		return
	}

	if len(data) < 1 {
		data, err = db.MongoClient.FindNewsByTitle(tag)
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "find news ok",
		"data":  data,
	})
}

// 半径千米
func findNewsByLocationHandler(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, "lat err")
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, "lon err")
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, "radis err")
		return
	}

	// 先查询ID
	locationIDs, err := db.RedisCli.GetLocationsInRadius(lat, lon, radius, 20)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Failed to get locations in radius"})
		return
	}

	data, err := db.MongoClient.FindNewsByNid(locationIDs)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": "Failed to get locations in radius"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "find news ok",
		"data":  data,
	})
}

// 设置计数
func addNewsFavCount(c *gin.Context) {
	nid := c.Query("nid") // 获取查询参数中的tag值
	uid := c.Query("uid")
	opt := c.Query("opt")
	if len(opt) == 0 {
		printErrResult(c, "fail", model.ErrBadParamCode, "opt must be inc or  dec")
		return
	}

	bInc := false
	if opt == "inc" {
		bInc = true
	} else if opt == "dec" {
		bInc = false
	} else {
		printErrResult(c, "fail", model.ErrBadParamCode, "opt must be add dec")
		return
	}

	db.MongoClient.AddNewsLikeFavHateCount(nid, uid, "fav", bInc, 0)

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "set fav news ok",
		"nid":   nid,
	})

}

func addNewsLikeCount(c *gin.Context) {
	nid := c.Query("nid") // 获取查询参数中的tag值
	uid := c.Query("uid")
	opt := c.Query("opt")
	if len(opt) == 0 {
		printErrResult(c, "fail", model.ErrBadParamCode, "opt must be inc or  dec")
		return
	}

	bInc := false
	if opt == "inc" {
		bInc = true
	} else if opt == "dec" {
		bInc = false
	} else {
		printErrResult(c, "fail", model.ErrBadParamCode, "opt must be add dec")
		return
	}

	db.MongoClient.AddNewsLikeFavHateCount(nid, uid, "like", bInc, 0)

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "set fav news ok",
		"nid":   nid,
	})
}

func addNewsHate(c *gin.Context) {
	nid := c.Query("nid") // 获取查询参数中的tag值
	uid := c.Query("uid")
	opt := c.Query("opt")
	reasonStr := c.Query("reason")
	reason, err := strconv.Atoi(reasonStr)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, "reason is a int param")
		return
	}

	bInc := false
	if opt == "inc" {
		bInc = true
	} else if opt == "dec" {
		bInc = false
	} else {
		printErrResult(c, "fail", model.ErrBadParamCode, "opt must be inc or  dec")
		return
	}

	db.MongoClient.AddNewsLikeFavHateCount(nid, uid, "hate", bInc, reason)

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "set fav news ok",
		"nid":   nid,
	})
}

// ////////////////////////////////////////////////////////////
// 添加文档的评论
func addNewsCommentHandler(c *gin.Context) {
	// 解析 JSON 数据到 Comment 结构体
	var comment model.Comment
	if err := c.BindJSON(&comment); err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, err.Error())
		return
	}

	uid := queryParamCommon(c, "uid")
	if comment.UID != uid {
		printErrResult(c, "fail", model.ErrBadParamCode, "uid not consist with uid in COMMENT json.")
		return
	}

	// 保存评论到 MongoDB
	if err := db.MongoClient.SaveNewsComment(&comment); err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "save comment ok",
		"cid":   comment.CID,
	})
}

// 查找评论
func findNewsCommentHandler(c *gin.Context) {
	// 从 URL 查询参数中获取页码和分页大小，默认为 1 和 10
	pageNumber, err := strconv.Atoi(c.Query("page"))
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}

	pageSize, err := strconv.Atoi(c.Query("size"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	// 查询 Comment 数据
	comments, err := db.MongoClient.GetCommentsSortedByTM(pageSize, pageNumber)
	if err != nil {
		printErrResult(c, "fail", http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "find comment ok",
		"data":  comments,
	})
}

// 删除一个文档
func deleteNewsCommentHandler(c *gin.Context) {
	nid := c.Query("nid") // 获取查询参数中的tag值
	cid := c.Query("cid") // 获取查询参数中的tag值

	if len(cid) == 0 || len(nid) == 0 {
		printErrResult(c, "fail", model.ErrBadParamCode, "nid and cid must be in")
		return
	}

	uid := queryParamCommon(c, "uid")
	// 在这里可以使用得到的 news 结构体进行后续处理
	err := db.MongoClient.DeleteCommentByMark(cid, nid, uid)
	if err != nil {
		common.Logger.Error("delete mongo err", zap.String("nid", nid))
	}

	// 应答 200
	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "delete comment ok",
		"nid":   nid,
	})
}

// 投诉与
func setReportHandler(c *gin.Context) {
	// 从请求 Body 中提取 JSON 数据
	body, err := c.GetRawData()
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, model.ErrBadParam.Error())
		return
	}

	// 调用 NewsFromJson 函数转为结构体
	news, err := model.NewsFromJsonBytes(body)
	if err != nil {
		printErrResult(c, "fail", model.ErrBadParamCode, model.ErrBadParam.Error())
		return
	}

	// 在这里可以使用得到的 news 结构体进行后续处理
	err = db.MongoClient.SaveNews(news)
	if err != nil {
		common.Logger.Error("save new to mongo err", zap.String("body", string(body)))
	}

	c.JSON(http.StatusOK, gin.H{
		"state": "ok",
		"code":  0,
		"des":   "i know, and will sort it later",
		"nid":   news.Nid,
	})
}
