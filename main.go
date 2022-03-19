package main

import (
	"YaeDisk/command/db"
	"YaeDisk/logx"
	"YaeDisk/router"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
	//_ "net/http/pprof"
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
	r.SetTrustedProxies(nil)
	r.MaxMultipartMemory = 256 << 20
	loadWebFile(r)
	router.Router(r)
	//go func() {
	//	log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	//}()
	r.Run(":9090")

	db.Close()
}
