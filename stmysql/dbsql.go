package stmysql

import (
	"database/sql"
	"fmt"
	"reflect"
)

func readRows(rows *sql.Rows, ttype reflect.Type) ([]interface{}, error) {
	cols, _ := rows.Columns()
	readCols := make([]interface{}, 0)
	for rows.Next() {
		colval := reflect.New(ttype).Interface()
		vals := reflect.ValueOf(colval).Elem()
		rets := make([]interface{}, 0, len(cols))
		for _, v := range cols {
			if f, ok := vals.Type().FieldByName(v); ok {
				rets = append(rets, vals.FieldByIndex(f.Index).Addr().Interface())
			} else {
				var i interface{}
				rets = append(rets, &i)
			}
		}
		err := rows.Scan(rets...)
		if err != nil {
			return readCols, err
		}
		readCols = append(readCols, colval)
	}

	return readCols, nil
}

func SelectOne(db *sql.DB, data interface{}, args ...interface{}) (int, error) {
	types := reflect.TypeOf(data)
	if types.Kind() != reflect.Ptr || types.Elem().Kind() != reflect.Struct {
		return 0, fmt.Errorf("[%s] you should give a ptr of a struct", types.Name())
	}
	rltype := types.Elem()
	sqlcmd := "select * from " + rltype.Name()
	if len(args) > 0 {
		if reflect.TypeOf(args[0]).Kind() != reflect.String {
			return 0, fmt.Errorf("[%s] args[0] should is a string", rltype.Name())
		}
		sqlcmd += " " + args[0].(string)
	}
	sqlcmd += " limit 1"

	var (
		rows *sql.Rows
		err  error
	)
	if len(args) > 1 {
		rows, err = db.Query(sqlcmd, args[1:]...)
	} else {
		rows, err = db.Query(sqlcmd)
	}
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cols, er := readRows(rows, rltype)
	if len(cols) > 0 {
		reflect.ValueOf(data).Elem().Set(reflect.ValueOf(cols[0]).Elem())
	}

	return len(cols), er
}

func SelectAll(db *sql.DB, data interface{}, args ...interface{}) ([]interface{}, error) {
	types := reflect.TypeOf(data)
	if types.Kind() != reflect.Ptr && types.Kind() != reflect.Struct {
		return nil, fmt.Errorf("[%s] you should give astruct or a ptr of a struct", types.Name())
	}
	if types.Kind() == reflect.Ptr && types.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("[%s] you should give astruct or a ptr of a struct", types.Name())
	}
	var rltype reflect.Type
	if types.Kind() == reflect.Struct {
		rltype = types
	} else {
		rltype = types.Elem()
	}
	sqlcmd := "select * from " + rltype.Name()
	if len(args) > 0 {
		if reflect.TypeOf(args[0]).Kind() != reflect.String {
			return nil, fmt.Errorf("[%s] args[0] should is a slice", rltype.Name())
		}
		sqlcmd += " " + args[0].(string)
	}

	var (
		rows *sql.Rows
		err  error
	)
	if len(args) > 1 {
		rows, err = db.Query(sqlcmd, args[1:]...)
	} else {
		rows, err = db.Query(sqlcmd)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readRows(rows, rltype)
}
