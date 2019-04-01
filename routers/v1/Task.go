package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"../../util"
	"strconv"
	"../../dbConnect"
	"strings"
	"database/sql"
	"../../exceptions"
	"time"
)

/**
* 创建任务
*/
func makeTask(context *gin.Context) {
	deviceIdsStr := context.PostForm("deviceIds")
	remark := context.PostForm("remark")
	if "" == deviceIdsStr {
		context.JSON(http.StatusBadRequest, util.Success("参数缺失"))
		return
	}
	strs := strings.Split(deviceIdsStr, ",")
	var ids = make([]int64, 0)
	for _, str := range strs {
		val, err := strconv.ParseInt(str, 10, 64)
		if nil != err {
			context.JSON(http.StatusBadRequest, util.Success("参数类型错误"))
			return
		}
		ids = append(ids, val)
	}
	taskId, err := dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("INSERT INTO task(createTime, remark, status) VALUES(?, ?, ?)")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		i, _ := stmt.Exec(time.Now().Unix(), remark, true)
		taskId, _ := i.LastInsertId()
		stmt.Close()
		stmt, err = tx.Prepare("INSERT INTO task_device_link(id, deviceId) VALUES(?, ?)")
		if nil != err {
			tx.Rollback()
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		for _, id := range ids {
			r, _ := stmt.Exec(taskId, id)
			row, _ := r.RowsAffected()
			if 0 == row {
				tx.Rollback()
				return nil, &exceptions.Error{Msg: "tasks insert fail.", Code: 500}
			}
		}
		return taskId, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(taskId))
}

/**
	查询巡检任务列表
 */
func getTaskList(context *gin.Context) {
	key := context.Query("key")
	if 32 != len(key) {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	reply, err := dbConnect.Select("device").
		Fields("id", "name", "remark", "ssid", "pwd").
		AND("status = ?", "task = ?").
		Query(true, key)
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
