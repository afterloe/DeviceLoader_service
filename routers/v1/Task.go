package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"../../util"
	"strconv"
	"../../dbConnect"
)

const (
	SUMMIT_SCENE = 1
	WUSHAN       = 2
)

/**
	查询巡检任务列表
 */
func getTaskList(context *gin.Context) {
	key := context.Param("key")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	reply, err := dbConnect.Select("device").
		Fields("id", "name", "remark", "ssid", "pwd").
		AND("status = ?", "task = ?").
		Query(true, val)
	if nil != err {
		context.JSON(http.StatusInternalServerError, util.Error(err))
		return
	}
	for _, item := range reply {
		id := item["id"].(int64)
		reply, _ := dbConnect.Select("warehouse").
			Fields("host", "url", "remarks", "lastSync").
			AND("device_id = ?", "status = ?").
			Query(id, true)
		item["points"] = reply
	}
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(reply))
}
