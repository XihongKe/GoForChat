package main

import (
	"GoForChat/helper"
	"GoForChat/ws"
	"github.com/gin-gonic/gin"
	"log"
)

var port = "8282"

//server
func main() {
	ipv4, err := helper.GetLocalIP()
	if err != nil {
		log.Fatalf("获取本机IP失败：%s", err)
	}
	log.Printf("服务运行在：%s:%s \n", ipv4, port)

	gin.SetMode(gin.DebugMode) //线上环境

	go ws.Manager.Start()
	r := gin.Default()
	r.GET("/ws", ws.UpgradeHandler)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("run time panic: %v", err)
		}
	}()
	_ = r.Run(":" + port) // listen and serve on 0.0.0.0:8080
}
