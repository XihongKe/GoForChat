package main

import (
	"GoForChat/helper"
	"GoForChat/ws"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"os"
)

var port = "8282"

//server
func main() {
	ipv4, err := helper.GetLocalIP()
	if err != nil {
		log.Fatalf("获取本机IP失败：%s", err)
	}
	log.Printf("LISTEN AND SERVE ON：%s:%s \n", ipv4, port)

	gin.SetMode(gin.DebugMode) //线上环境
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(os.Stdout, f)

	go ws.Manager.Start()
	r := gin.Default()
	r.GET("/ws", ws.UpgradeHandler)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("RUN TIME PANIC: %v", err)
		}
	}()
	_ = r.Run(":" + port)
}
