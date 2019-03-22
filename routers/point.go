package routers

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"net/http"
	"../util"
	"../dbConnect"
	"reflect"
	"../exceptions"
	"time"
	"database/sql"
	"../integrate/logger"
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

/**
	添加数据入口
*/
func newPoint(context *gin.Context) {
	key := context.PostForm("id")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}	
	p := &point{DeviceId: val, CreateTime: time.Now().Unix(), Status: true}
	p.Host = context.PostForm("host")
	p.Url = context.PostForm("url")
	p.Remarks = context.PostForm("remarks")
	err = p.Check("Host", "Url") // 参数检测
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("INSERT INTO warehouse(device_id, host, url, remarks, status, createTime) VALUES (?,?,?,?,?,?)")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, _ := stmt.Exec(p.DeviceId, p.Host, p.Url, p.Remarks, p.Status, p.CreateTime)
		id, _ := result.LastInsertId()
		p.Id = id
		logger.Logger("device", "insert success")
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success(p))
}

/**
	删除数据源
*/
func deletePoint(context *gin.Context) {
	key := context.PostForm("id")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("SELECT COUNT(1) FROM warehouse WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		c, _ := stmt.Query(val)
		c.Next()
		var count int64
		c.Scan(&count)
		if 0 == count {
			return nil, &exceptions.Error{Msg: "no such this point", Code: 404}
		}
		c.Close()
		stmt, err = tx.Prepare("UPDATE warehouse SET status = ? WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, _ := stmt.Exec(false, val)
		flag, _ := result.RowsAffected()
		if 0 == flag {
			return nil, &exceptions.Error{Msg: "delete fail", Code: 400}
		}
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success("delete success."))
}

/**
	修改数据源
*/
func modifyPoint(context *gin.Context) {
	key := context.PostForm("id")
	val, err := strconv.ParseInt(key, 10, 64)
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Fail(400, "参数错误"))
		return
	}
	_, err = dbConnect.WithTransaction(func(tx *sql.Tx) (interface{}, error) {
		stmt, err := tx.Prepare("SELECT device_id, host, url, remarks, lastSync FROM warehouse WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		c, _ := stmt.Query(val)
		c.Next()
		p := new(point)
		c.Scan(&p.DeviceId, &p.Host, &p.Url, &p.Remarks, &p.LastSync)
		c.Close()
		if 0 == p.DeviceId {
			return nil, &exceptions.Error{Msg: "no such this point", Code: 404}
		}
		flag := false
		if "" != context.PostForm("host") {
			p.Host = context.PostForm("host")
			flag = true
		}
		if "" != context.PostForm("url") {
			p.Url = context.PostForm("url")
			flag = true
		}
		if "" != context.PostForm("remarks") {
			p.Remarks = context.PostForm("remarks")
			flag = true
		}
		if false == flag {
			return nil, &exceptions.Error{Msg: "no change", Code: 400}
		}
		p.ModifyTime = time.Now().Unix()
		stmt, err = tx.Prepare("UPDATE warehouse SET host = ?, url = ?, remarks = ?, modifyTime = ? WHERE id = ?")
		if nil != err {
			return nil, &exceptions.Error{Msg: "db stmt open failed.", Code: 500}
		}
		result, _ := stmt.Exec(p.Host, p.Url, p.Remarks, p.ModifyTime, val)
		rows, _ := result.RowsAffected()
		if 0 == rows {
			return nil, &exceptions.Error{Msg: "update fail", Code: 400}
		}
		return nil, nil
	})
	if nil != err {
		context.JSON(http.StatusBadRequest, util.Error(err))
		return
	}
	context.JSON(http.StatusOK, util.Success("delete success."))
}

type point struct {
	DeviceId int64 `json:"deviceId"`
	Host string `json:"host"`
	Url string `json:"url"`
	Remarks string `json:"remarks"`
	Id int64 `json:"id"`
	Status bool `json:"status"`
	CreateTime int64 `json:"createTime"`
	ModifyTime int64 `json:"modifyTime"`
	LastSync int64 `json:"lastSync"`
}

/**
	参数检测
*/
func (this *point) Check(args ...string) error {
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