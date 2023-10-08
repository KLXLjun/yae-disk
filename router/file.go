package router

import (
	"YaeDisk/command/database"
	"YaeDisk/command/encrypt"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map/v2"
)

func FileRouter(file *gin.RouterGroup) {
	file.GET("/id/:id", func(c *gin.Context) {
		FileIDUParse := c.Param("id")
		if len(FileIDUParse) != 16 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "文件ID不正确",
			})
			return
		}

		crc32str := encrypt.Crc32Str(FileIDUParse)
		logx.Trace(path.Join(SavePath, crc32str[:4], FileIDUParse))
		if have, _ := utils.PathExists(path.Join(SavePath, crc32str[:4], FileIDUParse)); have {
			c.File(path.Join(SavePath, crc32str[:4], FileIDUParse))
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "文件不存在",
			})
		}
	})
	file.HEAD("/id/:id", func(c *gin.Context) {
		FileIDUParse := c.Param("id")
		if len(FileIDUParse) != 16 {
			c.Status(http.StatusBadRequest)
			return
		}

		crc32str := encrypt.Crc32Str(FileIDUParse)
		filePath := path.Join(SavePath, crc32str[:4], FileIDUParse)
		logx.Trace(filePath)
		if have, _ := utils.PathExists(filePath); have {
			get := utils.PathFileGet(filePath)
			if get != nil {
				c.Header("Content-Length", fmt.Sprintf("%d", get.Size()))
				c.Header("Content-Type", utils.GetContentType(filePath))
				c.Status(http.StatusOK)
				return
			}
			c.Status(http.StatusNotFound)
			return
		} else {
			c.Status(http.StatusNotFound)
		}
	})
	file.DELETE("/id/:id", func(c *gin.Context) {

	})

	file.POST("/path/*path", func(c *gin.Context) {
		Path := c.Param("path")
		pathArr := utils.PathProcess(strings.Split(Path, "/"))

		have, searchPath, _ := database.PathSearch(pathArr)
		if !have {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "没有这个文件夹",
			})
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
				fileID := utils.GenerateID()
				if fileID == "" {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "很抱歉,文件ID生成错误,请重试",
					})
					return
				}
				crc32str := encrypt.Crc32Str(fileID)

				logx.Debug("上传了文件", "文件夹:", Path, "名称:", files[0].Filename, "文件ID:", fileID)

				err := os.MkdirAll(TmpPath+"/"+crc32str[:4], os.ModePerm)
				if err != nil {
					logx.Warn("创建目录出错", TmpPath+"/"+crc32str[:4], err)
					c.JSON(http.StatusServiceUnavailable, gin.H{
						"error": "创建临时目录出错",
					})
					return
				}
				err = c.SaveUploadedFile(files[0], TmpPath+"/"+crc32str[:4]+"/"+fileID)
				filepath := SavePath + "/" + crc32str[:4] + "/" + fileID
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
					err = os.Rename(TmpPath+"/"+crc32str[:4]+"/"+fileID, filepath)
					if err != nil {
						c.JSON(http.StatusServiceUnavailable, gin.H{
							"error": "移动临时文件错误",
						})
						return
					}

					tm := time.Now()
					self := &model.FileRowStruct{
						ID:          fileID,
						IsDir:       false,
						OwnerFolder: searchPath.Self.ID,
						OwnerID:     "0",
						OwnerGroup:  "0",
						Name:        files[0].Filename,
						Size:        files[0].Size,
						ChangeTime:  tm,
						CreateTime:  tm,
					}
					if ok := database.Insert(self); ok == nil {
						searchPath.Child.Set(files[0].Filename, &database.File{
							Parent: searchPath,
							Self:   self,
							Child:  cmap.New[*database.File](),
						})
					} else {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": ok.Error(),
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
	file.HEAD("/path/*path", func(c *gin.Context) {
		Path := c.Param("path")
		pathArr := utils.PathProcess(strings.Split(Path, "/"))

		have, searchPath, _ := database.PathSearch(pathArr)
		if !have {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "路径未找到",
			})
		}

		if searchPath.Self.IsDir {
			c.Header("error", "这不是一个文件而是一个目录")
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		FileIDUParse := searchPath.Self.ID
		crc32str := encrypt.Crc32Str(FileIDUParse)
		filePath := path.Join(SavePath, crc32str[:4], FileIDUParse)
		logx.Trace(filePath)
		if have, _ := utils.PathExists(path.Join(SavePath, crc32str[:4], FileIDUParse)); have {
			get := utils.PathFileGet(filePath)
			if get != nil {
				c.Header("Accept-Ranges", "bytes")
				c.Header("Content-Length", fmt.Sprintf("%d", get.Size()))
				c.Header("Content-Type", utils.GetContentType(filePath))
				c.Header("Last-Modified", searchPath.Self.ChangeTime.Format(time.RFC3339))
				c.Header("Date", searchPath.Self.CreateTime.Format(time.RFC3339))
				c.Header("Ext-FileID", searchPath.Self.ID)
				c.Status(http.StatusOK)
			}
		} else {
			c.Header("error", "这个文件并不存在于本地磁盘!")
			c.Status(http.StatusNotFound)
		}
	})
	file.GET("/path/*path", func(c *gin.Context) {
		Path := c.Param("path")
		pathArr := utils.PathProcess(strings.Split(Path, "/"))
		have, searchPath, _ := database.PathSearch(pathArr)
		if !have {
			c.Header("error", "路径未找到")
			c.Status(http.StatusNotFound)
			return
		}

		if searchPath.Self.IsDir {
			c.Header("error", "这不是一个文件而是一个目录")
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		FileIDUParse := searchPath.Self.ID
		crc32str := encrypt.Crc32Str(FileIDUParse)
		logx.Trace(path.Join(SavePath, crc32str[:4], FileIDUParse))
		if have, _ := utils.PathExists(path.Join(SavePath, crc32str[:4], FileIDUParse)); have {
			c.File(path.Join(SavePath, crc32str[:4], FileIDUParse))
		} else {
			c.Header("error", "这个文件并不存在于本地磁盘!")
			c.Status(http.StatusNotFound)
		}
	})
	file.DELETE("/path/*path", func(c *gin.Context) {

	})
}
