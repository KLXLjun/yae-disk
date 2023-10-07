package main

import (
	"YaeDisk/command/database"
	"YaeDisk/logx"
	"YaeDisk/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	database.Start()
	database.LoadAll()

	logx.Info("main", "开始启动")
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.MaxMultipartMemory = 256 << 20
	loadWebFile(r)
	router.Router(r)

	srv := &http.Server{
		Addr:    "0.0.0.0:12800",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.Error("启动发生错误: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logx.Info("正在关闭服务器 ...")
}
