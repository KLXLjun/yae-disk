package tokenauth

import (
	"YaeDisk/logx"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TokenAuthServerHeader() func(c *gin.Context) {
	return func(c *gin.Context) {
		token, err2 := c.Cookie("token")
		if err2 != nil {
			logx.Trace("-403.1", "鉴权失败")
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.SetCookie("uid", "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "鉴权失败",
			})
			c.Abort()
			return
		}

		uid, err3 := c.Cookie("uid")
		if err3 != nil {
			logx.Trace("-403.2", "鉴权失败")
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.SetCookie("uid", "", -1, "/", "", false, true)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "鉴权失败",
			})
			c.Abort()
			return
		}

		var ok bool
		logx.Debug(token)

		if ok {
			logx.Debug(uid, "鉴权通过")
			c.Set("uid", uid)
			c.Next()
		} else {
			logx.Debug(uid, "鉴权失败")
			c.JSON(http.StatusForbidden, gin.H{
				"error": "鉴权失败",
			})
			c.Abort()
		}
	}
}
