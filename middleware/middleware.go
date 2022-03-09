package middleware

import (
	"YaeDisk/command/cache"
	"YaeDisk/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func AuthMiddleware(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"msg": "验证错误,请重新登录",
		})
		return
	}

	uid, err := c.Cookie("uid")
	if err != nil {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"msg": "验证错误,请重新登录",
		})
		return
	}

	parseUid, err := strconv.ParseUint(uid, 0, 64)
	if err != nil {
		return
	}

	ok, time := cache.CacheToken(parseUid, model.UserAuthToken{
		UserID:    parseUid,
		UserToken: token,
	})

	if ok {
		c.Set("time", time)
		c.Next()
	} else {
		c.JSON(http.StatusForbidden, map[string]interface{}{
			"msg": "验证错误,请重新登录",
		})
	}
}
