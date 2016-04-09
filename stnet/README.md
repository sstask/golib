stnet is a simple net lib.
example
### echo server
```
lster := ssnet.NewListener("127.0.0.1:6666", func(buf []byte) (parsedlen int, msg []byte) {
        return len(buf), buf
    }, func(sess *ssnet.Session, msg ssnet.SessionMsg) {
        fmt.Println(string(msg.Data))
        sess.Send(msg.Data)
    })
```
### echo client
```
conn := ssnet.NewConnector("127.0.0.1:6666", 100, func(buf []byte) (parsedlen int, msg []byte) {
        return len(buf), buf
    }, func(sess *ssnet.Session, msg ssnet.SessionMsg) {
        fmt.Println(string(msg.Data))
        sess.Send(msg.Data)
    })
```
