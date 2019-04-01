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

func modifyTask(context *gin.Context) {
	key := context.PostForm("key")
	deviceIdsStr := context.PostForm("deviceIds")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	if "" != deviceIdsStr {
		var ids = make([]int64, 0)
		strs := strings.Split(deviceIdsStr, ",")
		for _, str := range strs {
			val, err := strconv.ParseInt(str, 10, 64)
			if nil != err {
				context.JSON(http.StatusBadRequest, util.Success("参数类型错误"))
				return
			}
			ids = append(ids, val)
		}
		_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
			stmt, err := tx.Prepare("DELETE FROM task_device_link WHERE id = ?")
			if nil != err {
				tx.Rollback()
				return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
			}
			stmt.Exec(val)
			stmt.Close()
			stmt, err = tx.Prepare("INSERT INTO task_device_link(id, deviceId) VALUES(?, ?)")
			if nil != err {
				tx.Rollback()
				return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
			}
			for _, id := range ids {
				r, _ := stmt.Exec(val, id)
				row, _ := r.RowsAffected()
				if 0 == row {
					tx.Rollback()
					return nil, &exceptions.Error{Msg: "tasks insert fail.", Code: 500}
				}
			}
			return nil, nil
		})
	}
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	_, err = dbConnect.WithPrepare("UPDATE task SET remark = ? WHERE id = ?", func(stmt *sql.Stmt) (interface{}, error) {
		result, err := stmt.Exec(context.PostForm("remark"), val)
		if nil != err {
			return nil, &exceptions.Error{Msg: "db update failed.", Code: 500}
		}
		row, _ := result.RowsAffected()
		if 0 == row {
			return nil, &exceptions.Error{Msg: "tasks update fail.", Code: 500}
		}
		return nil, nil
	})

	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success("done"))
}

/**
* 创建任务
*/
func newTask(context *gin.Context) {
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
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	reply, err := dbConnect.Select("device").
		Fields("device.id", "device.name", "device.remark", "device.ssid", "device.pwd").
		JOIN("task_device_link").
		ON("task_device_link.deviceId = device.id").
		WHERE("task_device_link.id = ?").
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
