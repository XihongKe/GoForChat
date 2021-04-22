package ws

import (
	"GoForChat/helper"
	"GoForChat/user"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

//接收消息并进行判断处理的逻辑

// 消息类型
const (
	_ = iota
	MsgTypeUser
	MsgTypeGroup
	MsgTypeGetUser
	MsgTypeUserInfo
	MsgTypeSaveUserInfo
	MsgTypeGroupCreate
	MsgTypeUserLeave
)

// GroupMsgHandler 处理发给群组的消息
func GroupMsgHandler(request Request, messageStruct *Message) error {
	log.Printf("当前群组 %v", Manager.Groups)
	group, ok := Manager.Groups[messageStruct.Receiver]
	if !ok {
		return errors.New(fmt.Sprintf("消息接收者无效：%s", messageStruct.Receiver))
	}
	response := Message{
		Type:    MsgTypeGroup,
		Sender:  request.Client.ID,
		Content: messageStruct.Content,
		FromGroup: messageStruct.Receiver,
	}
	for k := range group.Member {
		conn, ok := Manager.Clients[group.Member[k].ID]
		if ok {
			response.Receiver = conn.ID
			responseJson, _ := json.Marshal(response)
			select {
			case conn.Send <- responseJson:
			default: // 超时未消费，释放相关资源
				close(conn.Send)
				delete(Manager.Clients, messageStruct.Receiver)
			}
		} else {
			log.Printf("群组存在无效组员：%v", group.Member[k])
		}
	}
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
	response, _ := json.Marshal(Message{
		Receiver: "",
		Type:     MsgTypeGetUser,
		Sender:   "SYSTEM",
		Content:  string(uJson),
	})
	request.Client.Send <- response
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
	request.Client.User.Name = fmt.Sprintf("%s#%s", userinfo.Name, helper.GetRandomString(6))
	request.Client.User.HeadImgUrl = userinfo.HeadImgUrl
	// 将用户的个人信息下发给客户端
	content, _ := json.Marshal(request.Client.User)
	response, _ := json.Marshal(Message{
		Receiver: request.Client.ID,
		Type:     MsgTypeUserInfo,
		Sender:   "SYSTEM",
		Content:  string(content),
	})
	request.Client.Send <- response
	usersBroadcast() // 广播用户列表
	return nil
}

// GroupCreateHandler 处理创建群组请求
func GroupCreateHandler(request Request, messageStruct *Message) error {
	type reqBody struct {
		MemberList []user.User `json:"member_list"`
		Name       string      `json:"name"`
	}
	var gb reqBody
	_ = json.Unmarshal([]byte(messageStruct.Content), &gb)
	group, err := user.NewGroup(request.Client.User, gb.MemberList, gb.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("创建群组失败：%s", err))
	}
	Manager.Groups[group.ID] = &group
	groupsBroadcast()
	return nil
}

// usersBroadcast 广播聊天列表
func usersBroadcast() {
	u := userList()
	uJson, _ := json.Marshal(u)

	// 生成消息并发送
	response, _ := json.Marshal(Message{
		Receiver: "",
		Type:     MsgTypeGetUser,
		Sender:   "SYSTEM",
		Content:  string(uJson),
	})
	broadcast(response)
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

// groupsBroadcast 群组列表广播
func groupsBroadcast() {
	for ID := range Manager.Clients {
		groups, _ := json.Marshal(groupList(ID))
		response, _ := json.Marshal(Message{
			Receiver: "",
			Type:     MsgTypeGroupCreate,
			Sender:   "SYSTEM",
			Content:  string(groups),
		})
		Manager.Clients[ID].Send <- response
	}
}

// groupList 获取用户参与的群组列表
func groupList(uid string) (gs []user.Group) {
	for k1 := range Manager.Groups {
		// 用户是群主
		if Manager.Groups[k1].Owner.ID == uid {
			gs = append(gs, *Manager.Groups[k1])
			continue
		}
		for _, m := range Manager.Groups[k1].Member {
			// 用户是群成员
			if m.ID == uid {
				gs = append(gs, *Manager.Groups[k1])
				break
			}
		}
	}
	return
}
