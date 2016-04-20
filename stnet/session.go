package stnet

import (
	"net"
	"sync/atomic"
)

//this will be called when session closed
type FuncOnClose func(*Session)

//message recv buffer size
const MsgBuffSize = 1024

//the length of send queue
const WriterListLen = 256

//session id
var GlobalSessionID uint64

type Session struct {
	MsgParse

	id     uint64
	socket net.Conn
	writer chan []byte
	closer chan int
	wclose chan int

	onclose FuncOnClose
}

func NewSession(con net.Conn, msgparse MsgParse, onclose FuncOnClose) *Session {
	if msgparse == nil {
		return nil
	}
	sess := &Session{
		id:       atomic.AddUint64(&GlobalSessionID, 1),
		socket:   con,
		writer:   make(chan []byte, WriterListLen),
		closer:   make(chan int),
		wclose:   make(chan int),
		MsgParse: msgparse,
		onclose:  onclose,
	}
	go sess.dosend()
	go sess.dorecv()
	return sess
}

func (this *Session) GetID() uint64 {
	return this.id
}

func (this *Session) Send(data []byte) bool {
	msg := make([]byte, len(data))
	copy(msg, data)
	select {
	case <-this.closer:
		return false
	case this.writer <- msg:
		return true
	}
}

func (this *Session) Close() {
	this.socket.Close()
}

func (this *Session) IsClose() bool {
	select {
	case <-this.closer:
		return true
	default:
		return false
	}
}

func (this *Session) dosend() {
	for {
		select {
		case <-this.wclose:
			goto exitsend
		case buf, ok := <-this.writer:
			if !ok {
				goto exitsend //chan closed
			}
			_, err := this.socket.Write(buf)
			if err != nil {
				this.socket.Close()
				goto exitsend
			}
		}
	}

exitsend:
	close(this.closer)
}

func (this *Session) dorecv() {
	this.ProcMsg(this, SessionMsg{CMD_NEW, nil})

	msgbuf := make([]byte, MsgBuffSize)
	msglen := 0
	for {
		if msglen*6/5 > len(msgbuf) {
			newbuf := make([]byte, len(msgbuf)*2)
			copy(newbuf, msgbuf)
			msgbuf = newbuf
		}
		buf := msgbuf[msglen:]
		n, err := this.socket.Read(buf)
		if err != nil {
			goto exitrecv
		}
		msglen += n
		dellen, msg := this.ParseMsg(msgbuf[0:msglen])
		if msg != nil {
			msgcpy := make([]byte, len(msg))
			copy(msgcpy, msg)
			this.ProcMsg(this, SessionMsg{CMD_DATA, msgcpy})
		}
		if dellen > 0 && dellen <= msglen {
			if dellen < msglen {
				copy(msgbuf, msgbuf[dellen:msglen])
			}
			msglen -= dellen
		}
	}

exitrecv:
	this.ProcMsg(this, SessionMsg{CMD_CLOSE, nil})
	close(this.wclose)
	this.socket.Close()
	<-this.closer //wait send routine exit
	if this.onclose != nil {
		this.onclose(this)
	}
}
