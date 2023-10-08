package router

import (
	"YaeDisk/command/database"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"net/http"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/gin-gonic/gin"
)

func FolderRouter(folder *gin.RouterGroup) {
	folder.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		pathArr := utils.PathProcess(strings.Split(path, "/"))
		logx.Debug("路径:"+path, pathArr, len(pathArr))
		ok, FolderInfo, AllFile := database.PathSearch(pathArr)
		if ok {
			c.JSON(http.StatusOK, gin.H{
				"folder_id":   FolderInfo.Self.ID,
				"folder_name": FolderInfo.Self.Name,
				"create_time": FolderInfo.Self.CreateTime,
				"change_time": FolderInfo.Self.ChangeTime,
				"file":        AllFile.File,
				"folder":      AllFile.Folder,
			})
		} else {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"msg":  "目录不存在",
				"path": path,
			})
		}
	})

	folder.POST("/:name/*path", func(context *gin.Context) {
		path := context.Param("path")
		FolderName := context.Param("name")
		pathArr := utils.PathProcess(strings.Split(path, "/"))

		have, searchPath, FolderAllFile := database.PathSearch(pathArr)
		if !have {
			context.JSON(http.StatusNotFound, gin.H{
				"error": "没有这个文件夹",
			})
		}

		for _, v := range FolderAllFile.File {
			if v.Name == FolderName {
				context.JSON(http.StatusBadRequest, gin.H{
					"error": "文件已存在",
				})
				return
			}
		}

		for _, v := range FolderAllFile.Folder {
			if v.Name == FolderName {
				context.JSON(http.StatusBadRequest, gin.H{
					"error": "文件夹已存在",
				})
				return
			}
		}

		fileID := utils.GenerateID()
		if fileID == "" {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": "很抱歉,文件ID生成错误,请重试",
			})
			return
		}

		tm := time.Now()
		self := &model.FileRowStruct{
			ID:          fileID,
			IsDir:       true,
			OwnerFolder: searchPath.Self.ID,
			OwnerID:     "0",
			OwnerGroup:  "0",
			Name:        FolderName,
			Size:        0,
			ChangeTime:  tm,
			CreateTime:  tm,
		}
		if ok := database.Insert(self); ok == nil {
			searchPath.Child.Set(FolderName, &database.File{
				Parent: searchPath,
				Self:   self,
				Child:  cmap.New[*database.File](),
			})
			context.Status(http.StatusOK)
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": ok.Error(),
			})
		}
	})
}
