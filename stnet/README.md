stnet is a simple net lib.
example
### echo server
```
lster,_ := stnet.NewListener("127.0.0.1:6666", &stnet.SimpleEchoMsgParse{})
```
### echo client
```
conn,_ := stnet.NewConnector("127.0.0.1:6666", 100, &stnet.SimpleEchoMsgParse{})

for {
		if conn.IsConnected() {
			fmt.Println(conn.InterData().(*st.SimpleEchoMsgParse).Mydata)
			conn.InterData().(*st.SimpleEchoMsgParse).Mydata = 1
			fmt.Println(conn.InterData().(*st.SimpleEchoMsgParse).Mydata)
			break
		}
	}
```
