stmysql is a lib to operate the database of mysql.
>
to create a new table or update a exiting table you can use "UpdateTableEx(dbTableDefine{})";
dbTableDefine is can be replaced by your struct;
also, you can use "CheckTableEx(dbTableDefine{})" to see the table's status in the database,
then use UpdateTableEx writing it into database;

example
```
type person struct {
	Id   uint
	Name string
	Age  uint8
}

tablestr, err1 := stmysql.UpdateTableEx(db, person{})
fmt.Println(tablestr, err1)
```