package ws

import (
	"GoForChat/helper"
	"GoForChat/user"
	"encoding/json"
	"errors"
	"fmt"
)

//接收消息并进行判断处理的逻辑

// GroupMsgHandler 处理发给群组的消息
func GroupMsgHandler(request Request, messageStruct *Message) error {
	return nil
}

// UserMsgHandler 处理发给特定用户的消息
func UserMsgHandler(request Request, messageStruct *Message) error {
	conn, ok := Manager.Clients[messageStruct.Receiver]
	if !ok {
		return errors.New(fmt.Sprintf("消息接收者无效：%s", messageStruct.Receiver))
	}
	select {
	case conn.Send <- request.Message:
	default: // 超时未消费，释放相关资源
		close(conn.Send)
		delete(Manager.Clients, messageStruct.Receiver)
	}
	return nil
}

// GetUserHandler 处理获取用户列表请求
func GetUserHandler(request Request, messageStruct *Message) error {
	u := userList()
	uJson, _ := json.Marshal(u)

	// 生成消息并发送
	response := Message{
		Receiver: "",
		Type:     MsgTypeGetUser,
		Sender:   "SYSTEM",
		Content:  string(uJson),
	}
	responseJson, _ := json.Marshal(response)
	request.Client.Send <- responseJson
	return nil
}

// SaveUserInfoHandler 处理客户端提交用户信息
func SaveUserInfoHandler(request Request, messageStruct *Message) error {
	userinfo := user.User{}
	_ = json.Unmarshal([]byte(messageStruct.Content), &userinfo)
	if userinfo.Name == "" || userinfo.HeadImgUrl == "" {
		return errors.New("用户信息缺失：" + string(request.Message))
	}
	// 设置用户信息
	//request.Client.User.Name = fmt.Sprintf("%s#%06v", userinfo.Name, rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000))
	request.Client.User.Name = fmt.Sprintf("%s#%s", userinfo.Name, helper.GetRandomString(6))
	request.Client.User.HeadImgUrl = userinfo.HeadImgUrl
	// 将用户的个人信息下发给客户端
	response, _ := json.Marshal(request.Client.User)
	message := Message{
		Receiver: request.Client.ID,
		Type:   MsgTypeUserInfo ,
		Sender:   "SYSTEM",
		Content:  string(response),
	}
	messageJson, _ := json.Marshal(message)
	request.Client.Send <- messageJson
	UserBroadcast() // 广播用户列表
	return nil
}

// UserBroadcast 广播用户列表
func UserBroadcast() {
	u := userList()
	uJson, _ := json.Marshal(u)

	// 生成消息并发送
	response := Message{
		Receiver: "",
		Type:     MsgTypeGetUser,
		Sender:   "SYSTEM",
		Content:  string(uJson),
	}
	responseJson, _ := json.Marshal(response)
	broadcast(responseJson)
}

// userList 获取在线用户列表
func userList() (u []user.User) {
	for k := range Manager.Clients {
		if Manager.Clients[k].User.Name == "" { // 过滤未设置用户信息的连接
			continue
		}
		u = append(u, Manager.Clients[k].User)
	}
	return
}

