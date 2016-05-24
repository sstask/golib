package stmysql

//in tag;
//mysql:"IsExport,Engine,Charset" key:PrimaryKey(..),Index(..),Index... autoinc:"ColumnName,StartNum";
//column:"IsExport,type";
//if mysql's IsExport is false, all columns's IsExport will be omited;
//these tags can be omited all;

type dbTableDefine struct {
	ID1  uint32 `mysql:"false,InnoDB,utf8" key:"primary(ID1,ID2),index(Name)" autoinc:"ID1,1"`
	ID2  uint32 `column:"true,INT(10) UNSIGNED"`
	Name string `column:"true,VARCHAR(255)"`
}

/*

 */
