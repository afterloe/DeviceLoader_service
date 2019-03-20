package routers

import (
	"github.com/gin-gonic/gin"
	"../dbConnect"
	"../util"
	"net/http"
	"strconv"
)

/**
	获取设备列表
*/
func getList(context *gin.Context)  {
	begin, limit:=pageCondition(context)
	reply, err := dbConnect.Select("device").
		Fields("id", "name", "remark").
		AND("status = ?").
		Page(begin, limit).
		Query(true)
	if nil != err {
		context.JSON(http.StatusInternalServerError, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(reply))
}

/**
	获取设备详情
 */
func getDevice(context *gin.Context) {
	key := context.Param("key")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	reply, err := dbConnect.Select("device").
		Fields("id", "name", "remark", "uid", "ssid", "pwd", "position", "createTime", "modifyTime").
		AND("id = ?", "status = ?").
		Query(val, true)
	if nil != err {
		context.JSON(http.StatusInternalServerError, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(reply))
}