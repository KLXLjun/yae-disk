//go:build debug
// +build debug

package main

import (
	"YaeDisk/logx"
	"github.com/gin-gonic/gin"
	"net/http"
)

func loadWebFile(Jinx *gin.Engine) {
	logx.Warn("正在以测试模式运行")

	Jinx.Static("/static", "web/static")
	Jinx.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/disk/")
	})

	Jinx.GET("/disk/*path", func(context *gin.Context) {
		//path := context.Param("path")
		context.File("./web/html/index.html")
	})
}
