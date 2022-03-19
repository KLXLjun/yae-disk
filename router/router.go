package router

import (
	"YaeDisk/command/db"
	"YaeDisk/command/encrypt"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var SavePath = "./save"

func Router(router *gin.Engine) {
	api := router.Group("/api")
	api.GET("/folder/create/:id/:name", func(context *gin.Context) {
		FolderIDUParse := context.Param("id")
		FolderName := context.Param("name")
		parseFID, err := strconv.ParseUint(FolderIDUParse, 10, 64)
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
	//api.GET("/file/uploadauth", func(context *gin.Context) {
	//
	//})
	api.POST("/file/upload/:id", func(c *gin.Context) {
		FolderIDUParse := c.Param("id")
		parseUint := utils.StringToUInt64(FolderIDUParse)
		if parseUint == math.MaxUint64 {
			logx.Warn("转换文件夹ID错误,上传文件无法继续")
			return
		}
		if !db.HaveFolder(parseUint) {
			logx.Warn("文件夹不存在,但仍然在上传", parseUint)
			return
		}
		form, err := c.MultipartForm()
		if err != nil {
			return
		}
		files := form.File["upload"]
		if len(files) > 0 {
			open, err := files[0].Open()
			if err != nil {
				logx.Warn("前端上传文件打开发生错误", err)
				err := open.Close()
				if err != nil {
					logx.Debug("无法关闭", err)
				}
				c.AbortWithStatus(http.StatusServiceUnavailable)
			} else {
				start1 := time.Now().Unix()
				crc32str := encrypt.Crc32SumFile(open)
				end1 := time.Now().Unix()
				end1time := end1 - start1

				start2 := time.Now().Unix()
				md5str := encrypt.Md5SumFile(open)
				end2 := time.Now().Unix()
				end2time := end2 - start2

				start3 := time.Now().Unix()
				sha1str := encrypt.Sha1SumFile(open)
				end3 := time.Now().Unix()
				end3time := end3 - start3

				start4 := time.Now().Unix()
				sha256str := encrypt.Sha256SumFile(open)
				end4 := time.Now().Unix()
				end4time := end4 - start4

				logx.Debug("上传了文件", "文件夹:", FolderIDUParse, "名称:", files[0].Filename, "crc32:", crc32str, end1time, "md5:", md5str, end2time, "sha1:", sha1str, end3time, "sha256:", sha256str, end4time)

				err := os.MkdirAll(SavePath+"/"+crc32str[:4], os.ModePerm)
				if err != nil {
					logx.Warn("创建目录出错", SavePath+"/"+crc32str[:4], err)
					return
				}
				err = c.SaveUploadedFile(files[0], SavePath+"/"+crc32str[:4]+"/"+sha256str)
				if err != nil {
					logx.Warn("前端上传文件保存错误", err)
				} else {
					tm := time.Now()
					nowtime := tm.Format("2006-01-02 15:04:05")
					db.InsFile(parseUint, model.FileStruct{
						FileID:     0,
						FolderID:   parseUint,
						FileName:   files[0].Filename,
						FileSize:   uint64(files[0].Size),
						ChangeTime: nowtime,
						CreateTime: nowtime,
						OwnerID:    0,
					})
				}

				err = open.Close()
				if err != nil {
					logx.Debug("无法关闭", err)
				}
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
				ChangeTime:    FolderInfo.ChangeTime,
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
