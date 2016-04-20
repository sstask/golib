stnet is a simple net lib.
example
### echo server
```
lster := stnet.NewListener("127.0.0.1:6666", stnet.SimpleEchoMsgParse{})
```
### echo client
```
conn := stnet.NewConnector("127.0.0.1:6666", 100, stnet.SimpleEchoMsgParse{})
```
