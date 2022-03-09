//go:build debug
// +build debug

package main

import (
	"YaeDisk/logx"
	"github.com/gin-gonic/gin"
)

func loadWebFile(Jinx *gin.Engine) {
	logx.Warn("正在以测试模式运行")

	Jinx.Static("/static", "web/static")
	Jinx.StaticFile("/", "./web/templates/index.html")
}
