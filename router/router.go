package router

import (
	"YaeDisk/command/db"
	"YaeDisk/logx"
	"YaeDisk/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func Router(router *gin.Engine) {
	router.GET("/disk/*path", func(context *gin.Context) {
		path := context.Param("path")
		context.String(http.StatusOK, "路径:"+path)
	})

	api := router.Group("/api")
	api.GET("/folder/create/:id/:name", func(context *gin.Context) {
		FolderIDUParse := context.Param("id")
		FolderName := context.Param("name")
		parseFID, err := strconv.ParseUint(FolderIDUParse, 0, 64)
		if err != nil {
			logx.Warn("转换文件夹ID出错 输入ID为:", FolderIDUParse, "错误为:", err)
			context.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"msg": "文件夹转换错误",
				"id":  FolderIDUParse,
			})
		} else {
			isOk, outFID, msg := db.InsFolder(parseFID, FolderName)
			if isOk {
				context.JSON(http.StatusOK, map[string]interface{}{
					"msg": "成功",
					"id":  outFID,
				})
			} else {
				context.JSON(http.StatusNotFound, map[string]interface{}{
					"msg": msg,
				})
			}
		}
	})

	api.GET("/disk/*path", func(context *gin.Context) {
		path := context.Param("path")
		pathArr := strings.Split(path, "/")
		xm := make([]string, 0)
		for _, s := range pathArr {
			if len(s) > 0 {
				xm = append(xm, s)
			}
		}
		logx.Debug("路径:"+path, xm, len(xm))
		ok, FolderInfo, FileList, FolderList := db.FolderList.PathSearch(xm)
		if ok {
			context.JSON(http.StatusOK, model.ResultFolderStruct{
				FolderID:      FolderInfo.FolderID,
				FolderName:    FolderInfo.FolderName,
				CreateTime:    FolderInfo.CreateTime,
				OwnerFolderID: FolderInfo.OwnerFolderID,
				OwnerUserID:   FolderInfo.OwnerUserID,
				File:          FileList,
				Folder:        FolderList,
			})
		} else {
			context.JSON(http.StatusNotFound, map[string]interface{}{
				"msg":  "目录不存在",
				"path": path,
			})
		}
	})
}
