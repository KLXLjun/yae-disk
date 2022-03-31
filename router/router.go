package router

import (
	"YaeDisk/command/db"
	"YaeDisk/command/encrypt"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var SavePath = "./save"
var TmpPath = "./tmp"

func Router(router *gin.Engine) {
	api := router.Group("/api")

	api.GET("/disk/*path", func(c *gin.Context) {
		path := c.Param("path")
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
			c.JSON(http.StatusOK, model.ResultFolderStruct{
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
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"msg":  "目录不存在",
				"path": path,
			})
		}
	})

	folder := api.Group("/folder")
	folder.POST("/:id/create/:name", func(context *gin.Context) {
		FolderIDUParse := context.Param("id")
		FolderName := context.Param("name")
		parseFID, err := strconv.ParseUint(FolderIDUParse, 10, 64)
		if err != nil {
			logx.Warn("转换文件夹ID出错 输入ID为:", FolderIDUParse, "错误为:", err)
			context.JSON(http.StatusBadRequest, gin.H{
				"error": "未知的文件夹ID",
			})
		} else {
			isOk, outFID := db.InsFolder(parseFID, FolderName)
			if isOk {
				context.Status(http.StatusOK)
			} else {
				if outFID == parseFID {
					context.JSON(http.StatusNotFound, gin.H{
						"error": "文件夹已存在",
					})
				} else {
					context.JSON(http.StatusNotFound, gin.H{
						"error": "创建文件夹时发生了错误",
					})
				}
			}
		}
	})

	file := api.Group("/file")
	file.POST("/:id", func(c *gin.Context) {
		uploadFileMD5 := c.GetHeader("filemd5")
		folderIDParse := c.Param("id")
		isOK, folderID := utils.StringToUInt64(folderIDParse)
		if !isOK {
			logx.Warn("转换文件夹ID错误,上传文件无法继续", "ID:", folderIDParse)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "未知的文件夹ID",
			})
			return
		}
		if !db.HaveFolder(folderID) {
			logx.Warn("文件夹不存在,但仍然在上传", folderID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "不存在的文件夹",
			})
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

				logx.Debug("上传了文件", "文件夹:", folderID, "名称:", files[0].Filename, "crc32:", crc32str, end1time, "md5:", md5str, end2time, "sha1:", sha1str, end3time, "sha256:", sha256str, end4time)

				if md5str != uploadFileMD5 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "文件验证错误",
					})
				}

				err := os.MkdirAll(TmpPath+"/"+crc32str[:4], os.ModePerm)
				if err != nil {
					logx.Warn("创建目录出错", TmpPath+"/"+crc32str[:4], err)
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error": "创建临时目录出错",
					})
					return
				}
				err = c.SaveUploadedFile(files[0], TmpPath+"/"+crc32str[:4]+"/"+sha256str)
				filepath := SavePath + "/" + crc32str[:4] + "/" + sha256str
				if err != nil {
					logx.Warn("前端上传文件保存错误", err)
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error": "上传文件保存错误",
					})
				} else {
					err := os.MkdirAll(SavePath+"/"+crc32str[:4], os.ModePerm)
					if err != nil {
						logx.Warn("创建存储目录出错", SavePath+"/"+crc32str[:4], err)
						c.JSON(http.StatusServiceUnavailable, gin.H{
							"error": "存储目录创建出错",
						})
						return
					}
					err = os.Rename(TmpPath+"/"+crc32str[:4]+"/"+sha256str, filepath)
					if err != nil {
						c.JSON(http.StatusServiceUnavailable, gin.H{
							"error": "移动临时文件错误",
						})
						return
					}

					tm := time.Now()
					nowTime := tm.Format("2006-01-02 15:04:05")
					isOK, errorMsg := db.InsFile(folderID, model.FileStruct{
						FileID:     0,
						FolderID:   folderID,
						FileName:   files[0].Filename,
						FileSize:   uint64(files[0].Size),
						ChangeTime: nowTime,
						CreateTime: nowTime,
						OwnerID:    0,
					}, filepath)
					if isOK {
						c.Status(http.StatusOK)
					} else {
						c.JSON(http.StatusServiceUnavailable, gin.H{
							"error": errorMsg,
						})
					}
				}

				err = open.Close()
				if err != nil {
					logx.Debug("无法关闭", err)
				}
			}
		}
	})
	file.GET("/:id", func(c *gin.Context) {
		FileIDUParse := c.Param("id")
		fileOK, fileID := utils.StringToUInt64(FileIDUParse)
		if !fileOK {
			logx.Warn("转换文件ID错误", "ID:", FileIDUParse)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "未知的文件ID",
			})
			return
		}

		if ok, val := db.GetFilePath(fileID); ok {
			c.File(val)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "文件不存在",
			})
		}
	})
	file.DELETE("/:id", func(c *gin.Context) {

	})
	file.GET("/:id/info", func(c *gin.Context) {

	})

	user := api.Group("/users")
	user.GET("/:id", func(c *gin.Context) {

	})

	userAuth := user.Group("/auth")
	userAuth.POST("/login", func(c *gin.Context) {
		name, nameHave := c.GetPostForm("name")
		pass, passHave := c.GetPostForm("pass")

		if !nameHave || !passHave {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "参数错误",
			})
		}

		isHave, User := db.GetUserFormName(name)
		if isHave {
			if User.Pass == pass {
				c.Status(http.StatusOK)
				return
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error": "密码错误",
			})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{
			"error": "用户不存在",
		})
		return
	})
}
