package router

import (
	"github.com/gin-gonic/gin"
)

var SavePath = "./save"
var TmpPath = "./tmp"

func Router(router *gin.Engine) {
	api := router.Group("/api")

	file := api.Group("/file")
	FileRouter(file)

	folder := api.Group("/folder")
	FolderRouter(folder)

	user := api.Group("/users")
	UserRouter(user)
}
