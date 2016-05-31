stnet is a simple net lib.
example
### echo server
```
lster := stnet.NewListener("127.0.0.1:6666", stnet.SimpleEchoMsgParse{})
```
### echo client
```
conn := stnet.NewConnector("127.0.0.1:6666", 100, stnet.SimpleEchoMsgParse{})

for {
		if conn.IsConnected() {
			fmt.Println(conn.Session.MsgParse.(*st.SimpleEchoMsgParse).Mydata)
			conn.Session.MsgParse.(*st.SimpleEchoMsgParse).Mydata = 1
			fmt.Println(conn.Session.MsgParse.(*st.SimpleEchoMsgParse).Mydata)
			break
		}
	}
```
