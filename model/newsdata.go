package model

import json "github.com/json-iterator/go"

/*
{
"nid": ""
"uid": "1001",
"nick": "",
"icon": "",
"lat": 40.0,
"log":116.0,
"alt":0.0,
"tm":0,
"title": "",
"content":"",
"images":["8767235711196729344.jpg", ""],
"tags":["颐和园", "桂花"],
"type": "point",  // track
"trackfile": "8767235711196729344.json",
"likes",
"favs",
}
*/
// Data 结构体表示给定的JSON数据
type News struct {
	Nid string `json:"nid"`

	Uid  string `json:"uid"`
	Nick string `json:"nick"`
	Icon string `json:"icon"`

	Lat float64 `json:"lat"`
	Log float64 `json:"log"`
	Alt float64 `json:"alt"`

	Tm        int64    `json:"tm"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Images    []string `json:"images"`
	Tags      []string `json:"tags"`
	Type      string   `json:"type"`
	TrackFile string   `json:"trackfile"`
	Likes     int      `json:"likes"`
	Favs      int      `json:"favs"`

	Deleted bool  `json:"deleted"`
	DelTm   int64 `json:"deltm"`
}

// GenerateData 生成 Data 结构体实例的函数
func NewNews() *News {
	return &News{
		Nid:       "",
		Uid:       "1000",
		Nick:      "",
		Icon:      "",
		Lat:       40.0,
		Log:       116.0,
		Alt:       0.0,
		Tm:        0,
		Title:     "",
		Content:   "",
		Images:    []string{},
		Tags:      []string{},
		Type:      "point",
		TrackFile: "",
		Likes:     0,
		Favs:      0,
		Deleted:   false,
		DelTm:     0,
	}
}

// SerializeData 将 Data 结构体序列化为 JSON 字符串
func (data *News) ToJson() (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// DeserializeData 将 JSON 字符串反序列化为 Data 结构体
func NewsFromJson(jsonData string) (*News, error) {
	var data News
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return &News{}, err
	}
	return &data, nil
}

func NewsFromJsonBytes(jsonData []byte) (*News, error) {
	var data News
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		return &News{}, err
	}
	return &data, nil
}

// ///////////////////////////////////////////////////////////////
// 评论
type Comment struct {
	NID string `json:"nid"`
	CID string `json:"cid"`

	UID  string `json:"uid"`
	Nick string `json:"nick"`
	Icon string `json:"icon"`

	PNID   string `json:"pnid"`
	TM     int64  `json:"tm"`
	ToID   string `json:"toid"`
	ToNick string `json:"tonick"`

	Content string   `json:"content"`
	Images  []string `json:"images"`
	Likes   int      `json:"likes"`
	Deleted bool     `json:"deleted"`
}

// NewComment 是 Comment 结构体的构造函数
func NewComment(nid, cid, uid, nick, icon, pNid, toID, toNick, content string, images []string, tm, likes int64) *Comment {
	return &Comment{
		NID:     nid,
		CID:     cid,
		UID:     uid,
		Nick:    nick,
		Icon:    icon,
		TM:      tm,
		PNID:    pNid,
		ToID:    toID,
		ToNick:  toNick,
		Content: content,
		Images:  images,
		Likes:   int(likes),
		Deleted: false,
	}
}

// ToJSON 将 Comment 结构体序列化为 JSON 格式的字符串
func (c *Comment) ToJSON() (string, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON 将 JSON 格式的字符串反序列化为 Comment 结构体
func CommentFromJSON(jsonData string) (*Comment, error) {
	data := Comment{}
	err := json.Unmarshal([]byte(jsonData), data)
	return &data, err
}

// 用户与文档的关联, 喜欢，收藏，讨厌拉黑，拉黑原因
type NewsFav struct {
	Nid    string `json:"nid"`
	Uid    string `json:"uid"`
	Like   bool   `json:"like"`
	Fav    bool   `json:"fav"`
	Hate   bool   `json:"hate"`
	Reason int    `json:"reason"`
}
