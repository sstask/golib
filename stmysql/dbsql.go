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
		structval := make(map[string]*[]byte)
		vals := reflect.ValueOf(colval).Elem()
		rets := make([]interface{}, 0, len(cols))
		for _, v := range cols {
			if f, ok := vals.Type().FieldByName(v); ok {
				val := vals.FieldByIndex(f.Index)
				if _, has := notEncodeTypes[reflect.TypeOf(val.Interface())]; !has {
					sli := make([]byte, 0)
					structval[v] = &sli
					rets = append(rets, &sli)
				} else {
					rets = append(rets, val.Addr().Interface())
				}
			} else {
				var i interface{}
				rets = append(rets, &i)
			}
		}
		err := rows.Scan(rets...)
		if err != nil {
			return readCols, err
		}
		for k, v := range structval {
			if f, ok := vals.Type().FieldByName(k); ok {
				val := vals.FieldByIndex(f.Index)
				err = unmarshal(*v, val.Addr().Interface())
				if err != nil {
					return readCols, err
				}
			}
		}
		readCols = append(readCols, colval)
	}

	return readCols, nil
}

func insertVal(slice []interface{}, val reflect.Value) ([]interface{}, error) {
	v := val.Interface()
	if _, has := notEncodeTypes[reflect.TypeOf(v)]; !has {
		m, err := marshal(v)
		if err != nil {
			return slice, err
		}
		return append(slice, m), nil
	}
	return append(slice, v), nil
}

func replaceORinsertOne(db *sql.DB, table interface{}, cmd string) (sql.Result, error) {
	types := reflect.TypeOf(table)
	if types.Kind() != reflect.Ptr && types.Kind() != reflect.Struct {
		return nil, fmt.Errorf("you should give a struct or a ptr of a struct")
	}
	if types.Kind() == reflect.Ptr && types.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("you should give a struct or a ptr of a struct")
	}
	rltype := types
	tbVal := reflect.ValueOf(table)
	if types.Kind() == reflect.Ptr {
		rltype = types.Elem()
		tbVal = tbVal.Elem()
	}
	if _, has := tableCfgType[rltype]; !has {
		_, err := addTable(table)
		if err != nil {
			return nil, err
		}
	}
	tbcfg, _ := tableCfgType[rltype]
	insertVals := make([]interface{}, 0, len(tbcfg.columns))
	sqlcmd := cmd + " into " + tbcfg.name + " ("
	sqlval := " values ("
	isfirst := true
	for _, v := range tbcfg.columns {
		if !v.export || v.autoinc {
			continue
		}
		if !isfirst {
			sqlcmd += ","
			sqlval += ","
		}
		sqlcmd += " " + v.name
		sqlval += " ?"
		isfirst = false

		var err error
		insertVals, err = insertVal(insertVals, tbVal.FieldByName(v.name))
		if err != nil {
			return nil, nil
		}
	}
	sqlcmd += ") " + sqlval + ")"

	return db.Exec(sqlcmd, insertVals...)
}

func replaceORinsertBatch(db *sql.DB, table interface{}, cmd string) (sql.Result, error) {
	types := reflect.TypeOf(table)
	if types.Kind() != reflect.Slice {
		return nil, fmt.Errorf("you should give a slice of a struct")
	}
	sliceval := reflect.ValueOf(table)
	if sliceval.Len() == 0 {
		return nil, nil
	}
	rltype := reflect.TypeOf(sliceval.Index(0).Interface())
	if rltype.Kind() == reflect.Ptr {
		rltype = rltype.Elem()
	}
	if _, has := tableCfgType[rltype]; !has {
		_, err := addTable(sliceval.Index(0).Interface())
		if err != nil {
			return nil, err
		}
	}
	tbcfg, _ := tableCfgType[rltype]
	insertVals := make([]interface{}, 0, len(tbcfg.columns)*sliceval.Len())
	insertCols := make([]string, 0, len(tbcfg.columns))
	sqlcmd := cmd + " into " + tbcfg.name + " ("
	isfirst := true
	for _, v := range tbcfg.columns {
		if !v.export || v.autoinc {
			continue
		}
		if !isfirst {
			sqlcmd += ","
		}
		sqlcmd += " " + v.name
		insertCols = append(insertCols, v.name)
		isfirst = false
	}
	sqlcmd += ") "

	sqlval := " values"
	isfirst = true
	for i := 0; i < sliceval.Len(); i++ {
		if !isfirst {
			sqlval += ","
		}
		sqlval += "("
		tbVal := reflect.ValueOf(sliceval.Index(i).Interface())
		if tbVal.Kind() == reflect.Ptr {
			tbVal = tbVal.Elem()
		}
		isf := true
		for _, v := range insertCols {
			if !isf {
				sqlval += ","
			}
			sqlval += " ?"
			isf = false
			var err error
			insertVals, err = insertVal(insertVals, tbVal.FieldByName(v))
			if err != nil {
				return nil, nil
			}
		}
		sqlval += ")"
		isfirst = false
	}

	sqlcmd += sqlval
	res, err := db.Exec(sqlcmd, insertVals...)
	return res, err
}

func SelectOne(db *sql.DB, table interface{}, args ...interface{}) (int, error) {
	types := reflect.TypeOf(table)
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
		reflect.ValueOf(table).Elem().Set(reflect.ValueOf(cols[0]).Elem())
	}

	return len(cols), er
}

func SelectAll(db *sql.DB, table interface{}, args ...interface{}) ([]interface{}, error) {
	types := reflect.TypeOf(table)
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

func GetRecordCount(db *sql.DB, table interface{}, args ...interface{}) (int, error) {
	tbname := ""
	ttype := reflect.TypeOf(table)
	if ttype.Kind() == reflect.String {
		tbname = table.(string)
	} else if ttype.Kind() == reflect.Struct {
		tbname = ttype.Name()
	} else if ttype.Kind() == reflect.Ptr && ttype.Elem().Kind() == reflect.Struct {
		tbname = ttype.Elem().Name()
	} else {
		return 0, fmt.Errorf("you should give a string or a struct or a ptr of struct")
	}
	sqlcmd := "select count(*) as num from " + tbname

	if len(args) > 0 {
		if reflect.TypeOf(args[0]).Kind() != reflect.String {
			return 0, fmt.Errorf("args[0] should is a string")
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
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		num := 0
		err = rows.Scan(&num)
		if err != nil {
			return 0, err
		}
		return num, nil
	}
	return 0, nil
}

func GetRecordMax(db *sql.DB, table interface{}, column string, args ...interface{}) (int, error) {
	tbname := ""
	ttype := reflect.TypeOf(table)
	if ttype.Kind() == reflect.String {
		tbname = table.(string)
	} else if ttype.Kind() == reflect.Struct {
		tbname = ttype.Name()
	} else if ttype.Kind() == reflect.Ptr && ttype.Elem().Kind() == reflect.Struct {
		tbname = ttype.Elem().Name()
	} else {
		return 0, fmt.Errorf("you should give a string or a struct as table")
	}
	sqlcmd := "select " + column + " as c from " + tbname

	if len(args) > 0 {
		if reflect.TypeOf(args[0]).Kind() != reflect.String {
			return 0, fmt.Errorf("args[0] should is a string")
		}
		sqlcmd += " " + args[0].(string)
	}
	sqlcmd += " order by c desc limit 1"

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

	for rows.Next() {
		max := 0
		err = rows.Scan(&max)
		if err != nil {
			return 0, err
		}
		return max, nil
	}
	return 0, nil
}

func InsertOne(db *sql.DB, table interface{}) (sql.Result, error) {
	return replaceORinsertOne(db, table, "insert")
}

func InsertBatch(db *sql.DB, table interface{}) (sql.Result, error) {
	return replaceORinsertBatch(db, table, "insert")
}

func ReplaceOne(db *sql.DB, table interface{}) (sql.Result, error) {
	return replaceORinsertOne(db, table, "replace")
}

func ReplaceBatch(db *sql.DB, table interface{}) (sql.Result, error) {
	return replaceORinsertBatch(db, table, "replace")
}

func DeleteRecord(db *sql.DB, table interface{}, args ...interface{}) (sql.Result, error) {
	tbname := ""
	ttype := reflect.TypeOf(table)
	if ttype.Kind() == reflect.String {
		tbname = table.(string)
	} else if ttype.Kind() == reflect.Struct {
		tbname = ttype.Name()
	} else if ttype.Kind() == reflect.Ptr && ttype.Elem().Kind() == reflect.Struct {
		tbname = ttype.Elem().Name()
	} else {
		return nil, fmt.Errorf("you should give a string or a struct or a ptr of struct")
	}
	sqlcmd := "delete from " + tbname

	if len(args) > 0 {
		if reflect.TypeOf(args[0]).Kind() != reflect.String {
			return nil, fmt.Errorf("args[0] should is a string")
		}
		sqlcmd += " " + args[0].(string)
	}

	if len(args) > 1 {
		return db.Exec(sqlcmd, args[1:]...)
	} else {
		return db.Exec(sqlcmd)
	}
}
