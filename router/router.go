package router

import (
	"YaeDisk/command/database"
	"YaeDisk/command/encrypt"
	"YaeDisk/command/utils"
	"YaeDisk/logx"
	"YaeDisk/model"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		ok, FolderInfo, AllFile := database.PathSearch(xm)
		if ok {
			c.JSON(http.StatusOK, gin.H{
				"folder_id":   FolderInfo.Self.ID,
				"folder_name": FolderInfo.Name,
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

	folder := api.Group("/folder")
	folder.POST("/create/:name/to/*path", func(context *gin.Context) {
		Path := context.Param("path")
		FolderName := context.Param("name")
		pathArr := strings.Split(Path, "/")
		xm := make([]string, 0)
		for _, s := range pathArr {
			if len(s) > 0 {
				xm = append(xm, s)
			}
		}

		have, searchPath, FolderAllFile := database.PathSearch(xm)
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
			searchPath.Child = append(searchPath.Child, &database.File{
				Name:   FolderName,
				IsDir:  true,
				Parent: searchPath,
				Self:   self,
				Child:  make([]*database.File, 0),
			})
			context.Status(http.StatusOK)
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": ok.Error(),
			})
		}
	})

	file := api.Group("/file")
	file.POST("/upload/to/*path", func(c *gin.Context) {
		Path := c.Param("path")
		pathArr := strings.Split(Path, "/")
		xm := make([]string, 0)
		for _, s := range pathArr {
			if len(s) > 0 {
				xm = append(xm, s)
			}
		}

		have, searchPath, _ := database.PathSearch(xm)
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
						searchPath.Child = append(searchPath.Child, &database.File{
							Name:   files[0].Filename,
							IsDir:  false,
							Parent: searchPath,
							Self:   self,
							Child:  nil,
						})
						c.Status(http.StatusOK)
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

	file.GET("/:id", func(c *gin.Context) {
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
	file.DELETE("/:id", func(c *gin.Context) {

	})
	file.GET("/:id/info", func(c *gin.Context) {

	})

	user := api.Group("/users")
	user.GET("/:id", func(c *gin.Context) {

	})

	userAuth := user.Group("/auth")
	userAuth.POST("/login", func(c *gin.Context) {
		//name, nameHave := c.GetPostForm("name")
		//pass, passHave := c.GetPostForm("pass")

		// if !nameHave || !passHave {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "参数错误",
		// 	})
		// }

		// isHave, User := db.GetUserFormName(name)
		// if isHave {
		// 	if User.Pass == pass {
		// 		c.Status(http.StatusOK)
		// 		return
		// 	}

		// 	c.JSON(http.StatusForbidden, gin.H{
		// 		"error": "密码错误",
		// 	})
		// 	return
		// }
		// c.JSON(http.StatusNotFound, gin.H{
		// 	"error": "用户不存在",
		// })
		return
	})
}
