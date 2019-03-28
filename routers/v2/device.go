package routers

import (
	"github.com/gin-gonic/gin"
	"../../dbConnect"
	"../../util"
	"net/http"
)

/**
	获取设备列表
*/
func getList(context *gin.Context)  {
	begin, limit:=pageCondition(context)
	reply, err := dbConnect.Select("device").
		Fields("id", "name", "position").
		AND("status = ?").
		Page(begin, limit).
		Query(true)
	if nil != err {
		context.JSON(http.StatusInternalServerError, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(reply))
}
