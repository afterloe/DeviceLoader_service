package main

import (
	"fmt"
	cli "./server" // 引入webService服务
	"runtime"
	"os"
	"./integrate/logger"
	"./config"
	"reflect"
)

var (
	defAddr string
	cpuNumber int
	pid int
)

func init() {
	defAddr = "0.0.0.0:8080"
	cpuNumber = runtime.NumCPU()
	pid = os.Getpid()
}

func multiServiceCfg(cfg map[string]interface{}) {
	if nil == cfg {
		logger.Error("main", "multi cpu config can't find.")
		return
	}
	flg := cfg["enable"]
	if nil == flg {
		logger.Info("main", "server not allow to use multi cpu")
		return
	}
	if flg.(bool) {
		num := config.GetByTarget(cfg, "num")
		var cpuNum = cpuNumber
		if nil != num {
			switch reflect.TypeOf(num).String() {
			case "float64":
				cpuNum = int(reflect.ValueOf(num).Float())
				break
			case "int":
				cpuNum = int(reflect.ValueOf(num).Int())
				break
			default:
				break
			}
		}
		if 0 >= cpuNum {
			cpuNum = cpuNumber
		}
		logger.Info("main", fmt.Sprintf("multi server model, server will use %v cpu", cpuNum))
		runtime.GOMAXPROCS(cpuNum * 2) // 限制go 出去的数量
	}
}

func main() {
	serverCfg := config.Get("server").(map[string]interface{})
	cfg := config.GetByTarget(serverCfg, "daemon").(map[string]interface{})
	addr, port := cfg["addr"], cfg["port"]
	if nil == addr {
		addr = "127.0.0.1"
	}
	if nil == port {
		port = "8080"
	}
	addrStr := fmt.Sprintf("%s:%s", addr, port)
	logger.Info("main",fmt.Sprintf("daemon server will start in %s ", addrStr))
	multiServiceCfg(serverCfg["multiCore"].(map[string]interface{}))
	cli.StartUpDaemonService(&defAddr, nil)
}