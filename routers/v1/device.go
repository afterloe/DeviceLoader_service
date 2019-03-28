package routers

import (
	"github.com/gin-gonic/gin"
	"../../dbConnect"
	"../../util"
	"net/http"
	"strconv"
	"time"
	"reflect"
	"../../exceptions"
	"database/sql"
	"../../integrate/logger"
	"fmt"
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
	context.JSON(http.StatusOK, util.Success(reply[0]))
}

/**
	创建设备
 */
func newDevice(context *gin.Context) {
	d := &device{
		Status: true,
		CreateTime: time.Now().Unix(),
	}
	d.Uid = context.PostForm("uid")
	d.Ssid = context.PostForm("ssid")
	d.Pwd = context.PostForm("pwd")
	d.Name = context.PostForm("name")
	d.Remark = context.PostForm("remark")
	d.Position = context.PostForm("position")
	err := d.Check("Name", "Ssid", "Pwd") // 参数检测
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("INSERT INTO device(uid, ssid, pwd, name, remark, position, status, createTime) VALUES (?,?,?,?,?,?,?,?)")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, _ := stmt.Exec(d.Uid, d.Ssid, d.Pwd, d.Name, d.Remark, d.Position, d.Status, d.CreateTime)
		id, _ := result.LastInsertId()
		d.Id = id
		logger.Logger("device", "insert success")
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(d))
}

/**
	修改设备
*/
func modifyDevice(context *gin.Context) {
	key := context.PostForm("key")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	dbStr := dbConnect.Select("device").
		Fields("id", "name", "remark", "status", "uid", "ssid", "pwd", "position", "createTime", "modifyTime").
		AND("id = ?").Preview()
	one, _:= dbConnect.WithQuery(dbStr, func(rows *sql.Rows) (interface{}, error) {
		target := new(device)
		flag := new(int64)
		rows.Next()
		rows.Scan(&target.Id, &target.Name, &target.Remark, &flag, &target.Uid, &target.Ssid, &target.Pwd, &target.Position, &target.CreateTime, &target.ModifyTime)
		if 1 == *flag {
			target.Status = true
		}
		rows.Close()
		return target, nil
	}, val)
	d := one.(*device)
	if 0 == d.Id {
		context.JSON(http.StatusBadRequest, util.Fail(404, "没有该设备"))
		return
	}
	flag := false
	if "" != context.PostForm("uid") {
		d.Uid = context.PostForm("uid")
		flag = true
	}
	if "" != context.PostForm("ssid") {
		d.Ssid = context.PostForm("ssid")
		flag = true
	}
	if "" != context.PostForm("pwd") {
		d.Pwd = context.PostForm("pwd")
		flag = true
	}
	if "" != context.PostForm("name") {
		d.Name = context.PostForm("name")
		flag = true
	}
	if "" != context.PostForm("remark") {
		d.Remark = context.PostForm("remark")
		flag = true
	}
	if "" != context.PostForm("position") {
		d.Position = context.PostForm("position")
		flag = true
	}
	if !flag {
		context.JSON(http.StatusBadRequest, util.Fail(400, "没有任何修改"))
		return
	}
	d.ModifyTime = time.Now().Unix()
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("UPDATE device SET uid = ?, ssid = ?, pwd = ?, name = ?, remark = ?, position = ?, status = ?, modifyTime = ? WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, err := stmt.Exec(d.Uid, d.Ssid, d.Pwd, d.Name, d.Remark, d.Position, d.Status, d.ModifyTime, d.Id)
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		rows, _ := result.RowsAffected()
		if 0 == rows {
			return nil, &exceptions.Error{Msg: "UPDATE FAILED.", Code: 500}
		}
		logger.Logger("device", fmt.Sprintf("rows -> %d", rows))
		logger.Logger("device", "update success")
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success("update success"))
}

/**
	删除设备
*/
func delDevice(context *gin.Context) {
	key := context.Param("key")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	dbStr := dbConnect.Select("device").
		Fields("id", "name", "remark", "status", "uid", "ssid", "pwd", "position", "createTime", "modifyTime").
		AND("id = ?").Preview()
	one, _:= dbConnect.WithQuery(dbStr, func(rows *sql.Rows) (interface{}, error) {
		target := new(device)
		flag := new(int64)
		rows.Next()
		rows.Scan(&target.Id, &target.Name, &target.Remark, &flag, &target.Uid, &target.Ssid, &target.Pwd, &target.Position, &target.CreateTime, &target.ModifyTime)
		if 1 == *flag {
			target.Status = true
		}
		rows.Close()
		return target, nil
	}, val)
	d := one.(*device)
	if 0 == d.Id {
		context.JSON(http.StatusBadRequest, util.Fail(404, "没有该设备"))
		return
	}
	if false == d.Status {
		context.JSON(http.StatusBadRequest, util.Fail(400, "该设备已被删除"))
		return
	}
	d.Status = false
	d.ModifyTime = time.Now().Unix()
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("UPDATE device SET uid = ?, ssid = ?, pwd = ?, name = ?, remark = ?, position = ?, status = ?, modifyTime = ? WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, err := stmt.Exec(d.Uid, d.Ssid, d.Pwd, d.Name, d.Remark, d.Position, d.Status, d.ModifyTime, d.Id)
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		rows, _ := result.RowsAffected()
		if 0 == rows {
			return nil, &exceptions.Error{Msg: "UPDATE FAILED.", Code: 500}
		}
		logger.Logger("device", fmt.Sprintf("rows -> %d", rows))
		logger.Logger("device", "update success")
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success("删除成功"))
}

type device struct {
	Id int64 `json:"id"`
	Uid string `json:"uid"`
	Ssid string `json:"ssid"`
	Pwd string `json:"pwd"`
	Name string `json:"name"`
	Remark string `json:"remark"`
	Position string `json:"position"`
	Status bool `json:"status"`
	CreateTime int64 `json:"createTime"`
	ModifyTime int64 `json:"modifyTime"`
}

/**
	参数检测
*/
func (this *device) Check(args ...string) error {
	value := reflect.ValueOf(*this)
	for _, arg := range args {
		v := value.FieldByName(arg)
		if !v.IsValid() {
			break
		}
		if "" == v.Interface() {
			return &exceptions.Error{Msg: "lack param " + arg, Code: 400}
		}
	}
	return nil
}