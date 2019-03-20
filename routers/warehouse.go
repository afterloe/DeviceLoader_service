package routers

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
	"../util"
	"../dbConnect"
)

/**
	获取数据入口
*/
func getPoint(context *gin.Context) {
	key := context.Param("key")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	// SELECT host, url, remarks, lastSync FROM warehouse WHERE warehouse.device_id = NULL AND warehouse.status = 'true';
	reply, err := dbConnect.Select("warehouse").
		Fields("host" ,"url", "remarks", "lastSync").
		AND("device_id = ?", "status = ?").
		Query(val, true)
	if nil != err {
		context.JSON(http.StatusInternalServerError, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(reply))
}
