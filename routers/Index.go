package routers

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
	"../util"
	"../config"
)

/**
	路由列表
 */
func Execute(route *gin.RouterGroup) {
	route.GET("/list", getList) // 获取设备列表
	route.GET("/device/:id", getDevice) // 获取设备详情
	route.POST("/device", modifyDevice) // 修改设备
	route.PUT("/device", newDevice) // 创建设备
	route.DELETE("/device/:id", delDevice) // 删除设备
	route.GET("/warehouse/:id", getPoint) // 获取数据入口
	route.PUT("/warehouse", newPoint) // 给设备添加数据源
	route.DELETE("/warehouse/:id", deletePoint) // 给设备删除数据源
	route.POST("/warehouse/:id") // 修改设备数据源
}

/**
	描述信息
 */
func Info(context *gin.Context) {
	info := config.Get("info").(map[string]interface{})
	context.JSON(http.StatusOK, util.Success(info))
}

/**
	分页组件
 */
func pageCondition(context *gin.Context) (int, int) {
	begin, err := strconv.Atoi(context.DefaultQuery("bg", "0"))
	if nil != err {
		begin = 0
	}
	end, err := strconv.Atoi(context.DefaultQuery("ed", "10"))
	if nil != err {
		end = 10
	}
	limit := end - begin
	if 0 >= limit {
		limit = 10
	}
	return begin, limit
}
