package middleware

import (
	"YaeDisk/command/cache"
	"YaeDisk/command/utils"
	"YaeDisk/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "验证错误,请重新登录",
		})
		return
	}

	uidCookie, err := c.Cookie("uid")
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "验证错误,请尝试重新登录解决问题",
		})
		return
	}

	isOk, UID := utils.StringToUInt64(uidCookie)
	if !isOk {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "验证错误,请尝试重新登录解决问题",
		})
		return
	}

	ip, ipOk := c.RemoteIP()
	if !ipOk {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "true",
			"msg":   "验证错误,请尝试重新登录解决问题",
		})
	}

	ok, time := cache.TokenCache(UID, model.UserAuthToken{
		UserID:    UID,
		UserToken: token,
		IP:        ip,
	})

	if ok {
		c.Set("time", time)
		c.Next()
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "true",
			"msg":   "验证错误,请重新登录",
		})
	}
}
