package stmysql

import (
	"reflect"
	"time"
)

type globalDBCfg struct {
	mysql string `engine:"InnoDB" charset:"utf8"`

	st_bool   bool        `type:"BOOL"`
	st_int8   int8        `type:"TINYINT(10)"`
	st_uint8  uint8       `type:"TINYINT(10) UNSIGNED"`
	st_int32  int32       `type:"INT(10)"`
	st_uint32 uint32      `type:"INT(10) UNSIGNED"`
	st_int64  int64       `type:"BIGINT(20)"`
	st_uint64 uint64      `type:"BIGINT(20) UNSIGNED"`
	st_float  float32     `type:"FLOAT(10)"`
	st_double float64     `type:"DOUBLE(20)"`
	st_string string      `type:"VARCHAR(255)"`
	st_bytes  []byte      `type:"BLOB(65535)"`
	st_time   time.Time   `type:"INT(10) UNSIGNED"`
	st_others interface{} `type:"BLOB(65535)"`
}

var MysqlCfg map[string]string
var DBColumnType map[reflect.Type]string

func init() {
	MysqlCfg = make(map[string]string)
	DBColumnType = make(map[reflect.Type]string)

	types := reflect.TypeOf(globalDBCfg{})
	glcfg, ok := types.FieldByName("mysql")
	if ok {
		MysqlCfg["engine"] = glcfg.Tag.Get("engine")
		MysqlCfg["charset"] = glcfg.Tag.Get("charset")
	}

	for i := 0; i < types.NumField(); i++ {
		val := types.Field(i).Tag.Get("type")
		if val != "" {
			DBColumnType[types.Field(i).Type] = val
		}
	}
}
