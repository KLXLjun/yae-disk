//go:build !debug
// +build !debug

package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:embed web/html/index.html
var indexPage []byte

//go:embed web/static
var static embed.FS

func loadWebFile(Jinx *gin.Engine) {
	Jinx.StaticFS("/static", http.FS(static))
	Jinx.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/disk/")
	})

	Jinx.GET("/disk/*path", func(context *gin.Context) {
		context.Data(http.StatusOK, "text/html;charset=utf-8", indexPage)
	})
}
