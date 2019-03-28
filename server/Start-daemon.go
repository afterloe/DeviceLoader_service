package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"../integrate/logger"
	"../integrate/notSupper"
	v1 "../routers/v1"
	v2 "../routers/v2"
	"os"
)

var notFoundStr, notSupperStr string

func init() {
	notFoundStr = "route is not defined."
	notSupperStr = "method is not supper"
}

/**
	启动守护进程

*/
func StartUpDaemonService(addr *string, cfg interface{}) {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	initDaemonService(engine, cfg)
	server := &http.Server{
		Addr: *addr,
		Handler: engine,
		MaxHeaderBytes: 1 << 20,
	}
	e := server.ListenAndServe()
	if nil != e {
		logger.Error("daemon", "server can't to run")
		logger.Error("daemon", e.Error())
		os.Exit(102)
	}
}

func initDaemonService(engine *gin.Engine, cfg interface{}) {
	engine.Use(gin.Recovery())
	engine.Use(logger.GinLogger())
	engine.Use(notSupper.HasError())
	engine.NoRoute(notSupper.NotFound(&notFoundStr))
	engine.NoMethod(notSupper.NotSupper(&notSupperStr))
	engine.MaxMultipartMemory = 120 << 20 // 最大上传文件为120 mb
	infoEntryPoint(engine)
	v1.Execute(engine.Group("/v1"))
	v2.Execute(engine.Group("/v2"))
	logger.Info("daemon", "daemon service is ready ...")
}

func infoEntryPoint(cxt *gin.Engine) {
	cxt.GET("/info", v1.Info)
}