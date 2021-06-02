package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"push_service/src/controllers"
	"push_service/src/core"
	"syscall"
)

func main() {
	//配置初始化
	core.Config.Init()
	//日志初始化
	core.Config.Logger.Init()
	//数据库初始化
	//models.Init()
	chExit := make(chan os.Signal)
	signal.Notify(chExit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	////客户端管理服务启动
	//go server.Manager.Start()
	////关闭超时连接客户端
	//go server.Manager.CloseTask()

	//监听http服务
	go func() {
		mux := http.NewServeMux()
		var mq controllers.GetMsgController
		var es controllers.GetEsController
		mux.HandleFunc("/MqConn", mq.MqConn)
		mux.HandleFunc("/CreateMapping", mq.CreateMapping)
		mux.HandleFunc("/SizeSearch", es.SizeSearch)
		log.Infof("Http Server started %s ...", core.Config.HttpListen)
		log.Fatal(http.ListenAndServe(core.Config.HttpListen, mux))
	}()

	//主进程阻塞直到有退出信号
	s := <-chExit
	log.Info("Get signal:", s)
}
