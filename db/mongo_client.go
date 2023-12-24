package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strconv"
	"time"
	"zhituBackend/common"
	"zhituBackend/model"
)

// go get go.mongodb.org/mongo-driver
const newsPrefix = "news"
const CommentPrefix = "comment"
const ReportPrefix = "report"
const UserFav = "userfav"

var MongoClient *MongoDBExporter = nil

func InitMongoClient(connectionString, dbName string) error {
	var err error
	MongoClient, err = NewMongoDBExporter(connectionString, dbName)

	return err
}

// MongoDBExporter 结构体
type MongoDBExporter struct {
	client *mongo.Client
	db     *mongo.Database
}

func getDateString(mt int64) string {
	t := time.UnixMilli(mt)
	// time.Now()
	return t.Format("20060102")
}

// NewMongoDBExporter 创建一个新的 MongoDBExporter 实例
func NewMongoDBExporter(connectionString, dbName string) (*MongoDBExporter, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	database := client.Database(dbName)

	return &MongoDBExporter{
		client: client,
		db:     database,
	}, nil
}

// close 关闭 MongoDBExporter
func (me *MongoDBExporter) Close() error {
	err := me.client.Disconnect(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func printTm(tm1, tm2 int64) {
	t1 := time.UnixMilli(tm1)
	t2 := time.UnixMilli(tm2)

	layout := "2006-01-02 15:04:05"

	fmt.Printf("%s---> %s  \n", t1.Format(layout), t2.Format(layout))
}

// init 初始化 MongoDBExporter
func (me *MongoDBExporter) Init(params map[string]string) error {
	// 在此进行初始化操作，如果有需要的话
	return nil
}

// exportMsg 导出消息到 MongoDB
// TODO: 需要加一个缓存，批量的写入，满100条，或者超时
func (me *MongoDBExporter) SaveNews(news *model.News) error {

	id := RedisCli.GetNextNewsId()
	news.Nid = strconv.FormatInt(id, 10)
	news.Tm = time.Now().UnixMilli()
	news.Deleted = false
	news.DelTm = 0

	collectionStatistic := me.db.Collection(newsPrefix)
	bsonData, err := bson.Marshal(news)
	_, err = collectionStatistic.InsertOne(context.Background(), bsonData)
	if err != nil {
		return err
	}

	// 保存都redis 中
	err = RedisCli.addGoeNews(news.Lat, news.Log, news.Nid)
	if err != nil {
		common.Logger.Error("save news to redis geo", zap.Error(err))
	}

	return nil
}

func (me *MongoDBExporter) DeleteOneNews(targetID string, uid string) error {
	coll := me.db.Collection(newsPrefix)
	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{"nid": targetID, "uid": uid}

	// 使用 DeleteOne 方法删除符合条件的第一个文档
	result, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		str := fmt.Sprintf("Deleted %v document(s)\n", result.DeletedCount)
		common.Logger.Info("Delete news: ", zap.String("info", str))

	}

	return nil
}

func (me *MongoDBExporter) DeleteOneNewsByMark(targetID string, uid string) error {
	coll := me.db.Collection(newsPrefix)
	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{"nid": targetID, "uid": uid}
	//filter := bson.M{"_id": targetID}

	// 构造更新操作，将 isDeleted 设置为 true
	update := bson.M{
		"$set": bson.M{"deleted": true, "deltm": time.Now().UnixMilli()},
	}

	// 使用 DeleteOne 方法删除符合条件的第一个文档
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount > 0 {
		str := fmt.Sprintf("Deleted %v document(s)\n", result.ModifiedCount)
		common.Logger.Info("Delete news: ", zap.String("info", str))

	}

	return nil
}

func (me *MongoDBExporter) DeleteManyNews(targetIDs []string) error {
	coll := me.db.Collection(newsPrefix)

	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{"nid": bson.M{"$in": targetIDs}}

	// 使用 DeleteMany 方法删除符合条件的所有文档
	result, err := coll.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		str := fmt.Sprintf("Deleted %v document(s)\n", result.DeletedCount)
		common.Logger.Info("Delete news: ", zap.String("info", str))
	}

	return nil
}

func (me *MongoDBExporter) DeleteManyNewsByMark(targetIDs []string) error {
	coll := me.db.Collection(newsPrefix)

	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{"nid": bson.M{"$in": targetIDs}}

	// 构造更新操作，将 isDeleted 设置为 true
	update := bson.M{
		"$set": bson.M{"deleted": true, "deltm": time.Now().UnixMilli()},
	}
	// 使用 DeleteMany 方法删除符合条件的所有文档
	result, err := coll.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount > 0 {
		str := fmt.Sprintf("Deleted %v document(s)\n", result.ModifiedCount)
		common.Logger.Info("Delete news: ", zap.String("info", str))
	}

	return nil
}

func (me *MongoDBExporter) findNewsFavByUidAndNid(nid, uid string) (*model.NewsFav, error) {
	coll := me.db.Collection(UserFav)
	// 构造查询条件
	filter := bson.M{"uid": uid, "nid": nid}

	// 执行查询
	var result model.NewsFav
	err := coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		// 处理错误
		return nil, err
	}

	// 返回查询结果
	return &result, nil
}

func (me *MongoDBExporter) addNewsFavByUidAndNid(nid, uid, field string, reason int) (*model.NewsFav, error) {
	coll := me.db.Collection(UserFav)

	bLike := field == "like"
	bFav := field == "fav"
	bHate := field == "hate"
	fav := model.NewsFav{Nid: nid, Uid: uid, Like: bLike, Fav: bFav, Hate: bHate, Reason: reason}
	bsonData, err := bson.Marshal(fav)
	_, err = coll.InsertOne(context.Background(), bsonData)
	if err != nil {
		return nil, err
	}

	// 返回查询结果
	return &fav, nil
}

func (me *MongoDBExporter) updateNewsFavByUidAndNid(nid, uid, field string, bSet bool, reason int) error {
	coll := me.db.Collection(UserFav)

	// 创建一个 filter 以匹配要更新的文档
	filter := bson.M{"nid": nid, "uid": uid}

	// 创建一个 update 文档，使用 $inc 操作符自增 likes 字段
	update := bson.M{
		"$set": bson.M{field: bSet},
	}

	if field == "hate" {
		update = bson.M{
			"$set": bson.M{field: bSet, "reason": reason},
		}
	}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	common.Logger.Info("update fav like hate", zap.String(nid+"-"+uid, field))

	return nil
}

func (me *MongoDBExporter) AddNewsLikeFavHateCount(nid, uid, field string, inc bool, reason int) error {

	fav, err := me.findNewsFavByUidAndNid(nid, uid)

	count := 0
	field1 := "likes"

	// 没有设置可以为
	if fav == nil {
		if inc {
			// 添加喜欢
			_, err = me.addNewsFavByUidAndNid(nid, uid, field, reason)
			if err != nil {
				return err
			}

			count = 1
			if "like" == field {
				field1 = "likes"
			} else if "fav" == field {
				field1 = "favs"
			}
		} else {
			// 取消喜欢，不需要操作，之前没有，也不应该执行到这里
		}
	} else { // 设置过，
		if inc {
			// 以前有，则需要检查该字符安是否设置了，没有设置字段才添加字段
			if "like" == field && fav.Like == false {
				me.updateNewsFavByUidAndNid(nid, uid, field, true, reason)
				field1 = "likes"
				count = 1
			} else if "fav" == field && fav.Fav == false {
				me.updateNewsFavByUidAndNid(nid, uid, field, true, reason)
				field1 = "favs"
				count = 1
			} else if "hate" == field && fav.Hate == false {
				me.updateNewsFavByUidAndNid(nid, uid, field, true, reason)
			}
		} else {
			// 取消喜欢，之前设置了才需要操作
			if "like" == field && fav.Like == true {
				me.updateNewsFavByUidAndNid(nid, uid, field, false, reason)
				count = -1
				field1 = "likes"
			} else if "fav" == field && fav.Fav == true {
				me.updateNewsFavByUidAndNid(nid, uid, field, false, reason)
				count = -1
				field1 = "favs"
			} else if "hate" == field && fav.Hate == true {
				me.updateNewsFavByUidAndNid(nid, uid, field, false, reason)
			}
		}
	}

	if count != 0 {
		err = me.addNewsFieldCount(nid, field1, count)
	}

	return err
}

// "like" "fav" +1 -1
func (me *MongoDBExporter) addNewsFieldCount(targetID string, field string, n int) error {
	coll := me.db.Collection(newsPrefix)

	// 创建一个 filter 以匹配要更新的文档
	filter := bson.M{"nid": targetID}

	// 创建一个 update 文档，使用 $inc 操作符自增 likes 字段
	update := bson.M{"$inc": bson.M{field: n}}

	// 使用 UpdateOne 方法更新符合条件的第一个文档
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.ModifiedCount > 0 {
		str := fmt.Sprintf("update %v document(s)\n", result.ModifiedCount)
		common.Logger.Info("update news: ", zap.String("info", str))
	}

	return nil
}

func (me *MongoDBExporter) FindLatestNews() ([]model.News, error) {
	coll := me.db.Collection(newsPrefix)

	// 按照 tm 字段降序排序，限制返回结果为20条
	findOptions := options.Find().SetSort(bson.D{{"tm", -1}}).SetLimit(20)

	// 构建查询过滤条件
	filter := bson.M{"deleted": false}

	// 执行查询
	cursor, err := coll.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解码结果
	var newsList []model.News
	err = cursor.All(context.Background(), &newsList)
	if err != nil {
		return nil, err
	}

	return newsList, nil
}

// 一次性查询多个文档，这个主要通过redis配合使用
func (me *MongoDBExporter) FindNewsByNid(nids []string) ([]model.News, error) {
	coll := me.db.Collection(newsPrefix)

	// 构建查询过滤条件
	filter := bson.M{"nid": bson.M{"$in": nids}}

	// 执行查询
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解码结果
	var newsList []model.News
	err = cursor.All(context.Background(), &newsList)
	if err != nil {
		return nil, err
	}

	return newsList, nil
}

// 生成单个的查询
func (me *MongoDBExporter) FindNewsByTag(tag string) ([]model.News, error) {
	coll := me.db.Collection(newsPrefix)

	// 构建查询过滤条件，使用 $regex 进行模糊匹配
	filter := bson.M{
		"tags":    bson.M{"$regex": primitive.Regex{Pattern: tag, Options: "i"}},
		"deleted": false,
	}

	// 执行查询
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解码结果
	var newsList []model.News
	err = cursor.All(context.Background(), &newsList)
	if err != nil {
		return nil, err
	}

	return newsList, nil
}

// 同时模糊查询
func (me *MongoDBExporter) FindNewsByTitle(keyword string) ([]model.News, error) {
	coll := me.db.Collection(newsPrefix)

	// 构建查询过滤条件，使用 $regex 进行模糊匹配
	// 构建查询过滤条件，使用 $or 操作符同时在 title 和 content 字段进行模糊匹配
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}},
			{"content": bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}},
		},
		"deleted": false,
	}

	// 执行查询
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 解码结果
	var newsList []model.News
	err = cursor.All(context.Background(), &newsList)
	if err != nil {
		return nil, err
	}

	return newsList, nil
}

//func main() {
//	// MongoDB 连接字符串，替换为你的实际连接字符串
//	connectionString := "mongodb://localhost:27017"
//
//	// 数据库名称
//	dbName := "your_database"
//
//	// 集合名称
//	collectionName := "your_collection"
//
//	// 创建一个新的 MongoDBExporter 实例
//	mongoDBExporter, err := NewMongoDBExporter(connectionString, dbName, collectionName)
//	if err != nil {
//		log.Fatal("Error creating MongoDB exporter:", err)
//	}
//
//	// 初始化
//	if err := mongoDBExporter.init(nil); err != nil {
//		log.Fatal("Initialization failed:", err)
//	}
//
//	// 模拟导出消息到 MongoDB
//	for i := 1; i <= 5; i++ {
//		msg := fmt.Sprintf("Message %d", i)
//		if err := mongoDBExporter.exportMsg(msg); err != nil {
//			log.Fatal("Exporting message failed:", err)
//		}
//
//		time.Sleep(time.Second)
//	}
//
//	// 关闭 MongoDBExporter
//	if err := mongoDBExporter.close(); err != nil {
//		log.Fatal("Closing exporter failed:", err)
//	}
//}

func (me *MongoDBExporter) SaveNewsComment(com *model.Comment) error {

	id := RedisCli.GetNextCommentId()
	com.CID = strconv.FormatInt(id, 10)
	com.TM = time.Now().UnixMilli()

	coll := me.db.Collection(CommentPrefix)
	bsonData, err := bson.Marshal(com)
	_, err = coll.InsertOne(context.Background(), bsonData)
	if err != nil {
		return err
	}

	return nil
}

func (me *MongoDBExporter) DeleteComment(cid, nid, uid string) error {
	coll := me.db.Collection(CommentPrefix)

	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{
		"nid": nid,
		"cid": cid,
		"uid": uid,
	}
	// 使用 DeleteOne 方法删除符合条件的第一个文档
	result, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount > 0 {
		str := fmt.Sprintf("Deleted %v comment(s)\n", result.DeletedCount)
		common.Logger.Info("Delete comment: ", zap.String("info", str))

	}

	return nil
}

func (me *MongoDBExporter) DeleteCommentByMark(cid, nid, uid string) error {
	coll := me.db.Collection(CommentPrefix)

	// 创建一个 filter 以匹配要删除的文档
	filter := bson.M{
		"nid": nid,
		"cid": cid,
		"uid": uid,
	}

	// 构造更新操作，将 isDeleted 设置为 true
	update := bson.M{
		"$set": bson.M{"deleted": true},
	}
	// 使用 DeleteMany 方法删除符合条件的所有文档
	result, err := coll.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	if result.ModifiedCount > 0 {
		str := fmt.Sprintf("Deleted %v comment(s)\n", result.ModifiedCount)
		common.Logger.Info("Delete comment: ", zap.String("info", str))

	}

	return nil
}

// GetCommentsSortedByTM 分页查询并根据 tm 字段排序的评论
func (me *MongoDBExporter) GetCommentsSortedByTM(pageSize, pageNumber int) ([]model.Comment, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"tm", -1}}) // 根据 tm 字段降序排序

	// 计算跳过的文档数，实现分页效果
	skip := pageSize * (pageNumber - 1)
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))

	coll := me.db.Collection(CommentPrefix)
	cursor, err := coll.Find(ctx, bson.M{"deleted": false}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []model.Comment
	for cursor.Next(ctx) {
		var comment model.Comment
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (me *MongoDBExporter) SaveReport(news *model.News) error {

	id := RedisCli.GetNextNewsId()
	news.Nid = strconv.FormatInt(id, 10)
	news.Tm = time.Now().UnixMilli()

	coll := me.db.Collection(ReportPrefix)
	bsonData, err := bson.Marshal(news)
	_, err = coll.InsertOne(context.Background(), bsonData)
	if err != nil {
		return err
	}

	return nil
}
