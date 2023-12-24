package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

func InitWebSocket(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			fmt.Println("升级协议", r.Header["User-Agent"])
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	userId := c.Query("id")
	fmt.Println("用户id:", userId)

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 连接关闭
				fmt.Println("连接关闭:", err)
			} else {
				// 其他错误
				fmt.Println("读错误:", err)
			}
			break
		}

		fmt.Println("获取客户端发送的消息:" + string(message))
		fmt.Println("类型：", mt)

		var msg = "春风得意马蹄疾,一日看尽长安花"
		err2 := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err2 != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 连接关闭
				fmt.Println("连接关闭:", err2)
			} else {
				// 其他错误
				fmt.Println("写错误:", err2)
			}
			break
		}
	}

}
