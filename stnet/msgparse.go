package stnet

import (
	"fmt"
)

const (
	CMD_NEW int = iota
	CMD_CLOSE
	CMD_DATA
)

type SessionMsg struct {
	Cmd  int    `messagetype`
	Data []byte `messagedata`
}

type MsgParse interface {
	ParseMsg(buf []byte) (parsedlen int, msg []byte)
	//buf:recved data now;
	//parsedlen:length of recved data parsed;
	//msg: message which is parsed from recved data

	ProcMsg(*Session, SessionMsg)
	//*Session:session which recved message
	//SessionMsg: message recved
}

//a simple example of "MsgParse interface"
//it's a echo server or client,when it recv a message, it send the message back at once
type SimpleEchoMsgParse struct {
	//you can put any data here,and use Session to access it
	Mydata int
}

func (SimpleEchoMsgParse) ParseMsg(buf []byte) (parsedlen int, msg []byte) {
	return len(buf), buf
}
func (SimpleEchoMsgParse) ProcMsg(sess *Session, msg SessionMsg) {
	if msg.Cmd == CMD_NEW {
		fmt.Println("new socket ", sess.GetID())
		msg := []byte("hello")
		sess.Send(msg)
	} else if msg.Cmd == CMD_CLOSE {
		fmt.Println("socket closed ", sess.GetID())
	} else if msg.Cmd == CMD_DATA {
		fmt.Printf("socket %d recv msg:	%s\n", sess.GetID(), string(msg.Data))
		sess.Send(msg.Data)
	}
}
