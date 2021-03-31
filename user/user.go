package user

import (
	"errors"
	"fmt"
	"github.com/bxcodec/faker/v3"
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
		ID:         faker.UUIDDigit(),
		Name:       "",
		HeadImgUrl: "/img/default.png",
	}
}

// Group 群组
type Group struct {
	ID string
	Name string
	Member []User
}

// NewGroup 返回一个新群组
func NewGroup(users []User) Group{
	return Group{
		ID: faker.UUIDDigit(),
		Name: faker.Username(),
		Member: users,
	}
}

// FindGroup 查找特定群组
func FindGroup(ID string) (Group, error) {
	return Group{}, errors.New("group not found")
}
