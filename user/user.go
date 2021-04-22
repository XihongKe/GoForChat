package user

import (
	"errors"
	"fmt"
	"github.com/rs/xid"
)

// User 用户对象
type User struct {
	ID         string `json:"ID"`
	Name       string `json:"name"`
	HeadImgUrl string `json:"head_img_url"`
}

func (user User) String() string {
	return fmt.Sprintf("ID: %s, Name: %s, HeaderImgUrl: %s", user.ID, user.Name, user.HeadImgUrl)
}

// NewUser 返回一个新用户
func NewUser() User {
	return User{
		ID:         xid.New().String(),
		Name:       "",
		HeadImgUrl: "/img/default.png",
	}
}

// Group 群组
type Group struct {
	ID        string `json:"ID"`
	Owner     User `json:"owner"`
	Name      string `json:"name"`
	Member    []User `json:"member"`
	LeaveChan chan string `json:"-"` // 用户离开队列
}

// NewGroup 返回一个新群组
func NewGroup(owner User, users []User, name string) (g Group, e error) {
	if len(users) < 3 {
		return g, errors.New("需2个以上成员")
	}
	if name == "" {
		name = fmt.Sprintf("%s、%s等%d名成员", users[0].Name, users[1].Name, len(users))
	}
	g = Group{
		ID:        xid.New().String(),
		Owner:     owner,
		Name:      name,
		Member:    users,
		LeaveChan: make(chan string),
	}
	go g.UserLeave()
	return
}

// UserLeave 从群组中删除一个用户
func (g *Group) UserLeave() {
	defer func() {
		close(g.LeaveChan)
	}()

	for {
		uid, ok := <-g.LeaveChan
		if !ok {
			return
		}

		for k := range g.Member {
			if g.Member[k].ID != uid {
				continue
			}
			// 移除对应用户
			g.Member = append(g.Member[:k], g.Member[k+1:]...)
			break
		}
	}
}

// FindGroup 查找特定群组
func FindGroup(ID string) (Group, error) {
	return Group{}, errors.New("group not found")
}
