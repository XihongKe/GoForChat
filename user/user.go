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
	ID     string
	Owner  User
	Name   string
	Member []User
}

// NewGroup 返回一个新群组
func NewGroup(owner User, users []User, name string) (Group, error) {
	if len(users) < 3 {
		return Group{}, errors.New("需2个以上成员")
	}
	if name == "" {
		name = fmt.Sprintf("%s，%s等%d名成员", users[0].Name, users[1].Name, len(users))
	}
	return Group{
		ID:     xid.New().String(),
		Owner:  owner,
		Name:   name,
		Member: users,
	}, nil
}

// FindGroup 查找特定群组
func FindGroup(ID string) (Group, error) {
	return Group{}, errors.New("group not found")
}
