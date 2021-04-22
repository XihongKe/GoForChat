package ws

import (
	"GoForChat/user"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// ClientManager 客户端管理器
type ClientManager struct {
	Clients    map[string]*Client
	Groups     map[string]*user.Group
	Broadcast  chan Request
	Register   chan *Client
	Unregister chan *Client
}

// Client websocket客户端
type Client struct {
	ID     string
	Socket *websocket.Conn
	Send   chan []byte
	User   user.User
}

// Message 是通信的消息
type Message struct {
	Receiver  string `json:"receiver,omitempty"`
	Type      int    `json:"type,omitempty"`
	Sender    string `json:"sender,omitempty"`
	FromGroup string `json:"from_group,omitempty"` // 来自群聊的id，如果是私聊则为空
	Content   string `json:"content,omitempty"`
}

// Request 是请求包 包含客户端连接，及JSON格式的请求消息
type Request struct {
	Client *Client
	Message []byte
}

// Manager 声明管理器
var Manager = ClientManager{
	Broadcast:  make(chan Request),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[string]*Client),
	Groups: 	make(map[string]*user.Group),
}

/**
userA  userB  userC  userD
|		|		|		|
wsA	   wsB	   wsC	  wsD
|	?	|
- ------
		server

*/
// Start 项目运行前, 协程开启start -> go Manager.Start()
func (manager *ClientManager) Start() {
	for {
		select {
		case client := <-Manager.Register: // 新用户加入队列
			log.Printf("新用户加入:%v", client.ID)
			Manager.Clients[client.ID] = client

		case client := <-Manager.Unregister: // 用户离开队列
			log.Printf("用户离开:%v", client.ID)
			if _, ok := Manager.Clients[client.ID]; ok {
				// 广播用户离开的信息
				m, _ := json.Marshal(Message{
					Receiver: "",
					Type:     MsgTypeUserLeave,
					Sender:   "SYSTEM",
					Content:  client.ID,
				})
				broadcast(m)
				// 删除群组中成员
				for _, group := range groupList(client.ID){
					group.LeaveChan <- client.ID
				}
				// 资源销毁
				close(client.Send)
				delete(Manager.Clients, client.ID)
			}


		case request := <-Manager.Broadcast: // 处理消息发送
			MessageStruct := Message{}
			_ = json.Unmarshal(request.Message, &MessageStruct)
			switch MessageStruct.Type {
			case MsgTypeGroup: // 发送给群组
				err := GroupMsgHandler(request, &MessageStruct)
				if err != nil {
					log.Printf("GroupMsgHandler错误：%v", err)
				}
			case MsgTypeUser: // 发送给单个用户
				err := UserMsgHandler(request, &MessageStruct)
				if err != nil {
					log.Printf("UserMsgHandler错误：%v", err)
				}
			case MsgTypeGetUser: // 获取用户列表
				err := GetUserHandler(request, &MessageStruct)
				if err != nil {
					log.Printf("GetUserHandler错误：%v", err)
				}
			case MsgTypeSaveUserInfo: // 提交用户信息
				err := SaveUserInfoHandler(request, &MessageStruct)
				if err != nil {
					log.Printf("SaveUserInfoHandler错误：%v", err)
				}
			case MsgTypeGroupCreate:
				err := GroupCreateHandler(request, &MessageStruct)
				if err != nil {
					log.Printf("GroupCreateHandler错误：%s", err)
				}
			default:
				log.Printf("无效的Type类型：%s", request.Message)
			}

		}
	}
}

func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
		log.Printf("%s离开,关闭读取", c.ID)
	}()

	for {
		c.Socket.PongHandler()
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			log.Printf("读取客户端消息失败：%s", err)
			break
		}
		//log.Printf("读取到客户端的信息:%s", string(message))
		// 消息加入广播队列
		Manager.Broadcast <- Request{c, message}
	}
}

func (c *Client) Write() {
	defer func() {
		log.Printf("%s离开,关闭写入", c.ID)
		_ = c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//log.Printf("发送到到客户端的信息:%s", string(message))

			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Close 关闭连接 处理用户离开等逻辑
func (c *Client) Close(){

	close(c.Send)
	delete(Manager.Clients, c.ID)
}

//UpgradeHandler 处理创建用户、启动读写协程等逻辑
func UpgradeHandler(c *gin.Context) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	newUser := user.NewUser()
	client := &Client{
		ID:     newUser.ID,
		Socket: conn,
		Send:   make(chan []byte),
		User:   newUser,
	}
	Manager.Register <- client
	log.Printf("创建了新用户 %s", newUser)
	go client.Read()
	go client.Write()
}

// broadcast 向所有客户端广播消息
func broadcast(message []byte){
	for _, conn := range Manager.Clients {
		conn.Send <- message
	}
}