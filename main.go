package main

import (
	"YaeDisk/command/db"
	"YaeDisk/logx"
	"YaeDisk/router"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		TimestampFormat: time.RFC3339,
		FieldsOrder:     []string{"func"},
	})
	logx.Info("run")
	db.InitDB()
	logx.Info("main", "开始启动")
	r := gin.Default()
	loadWebFile(r)
	router.Router(r)
	r.Run(":9090")

	db.Close()
}
