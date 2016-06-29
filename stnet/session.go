package stnet

import (
	"fmt"
	"net"
	"reflect"
	"sync/atomic"
	"time"
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

func NewSession(con net.Conn, msgparse MsgParse, onclose FuncOnClose) (*Session, error) {
	if msgparse == nil {
		return nil, fmt.Errorf("MsgParse should not be nil")
	}
	sess := &Session{
		id:       atomic.AddUint64(&GlobalSessionID, 1),
		socket:   con,
		writer:   make(chan []byte, WriterListLen), //It's OK to leave a Go channel open forever and never close it. When the channel is no longer used, it will be garbage collected.
		closer:   make(chan int),
		wclose:   make(chan int),
		MsgParse: reflect.New(reflect.TypeOf(msgparse)).Interface().(MsgParse),
		onclose:  onclose,
	}
	go sess.dosend()
	go sess.dorecv()
	return sess, nil
}

func (this *Session) GetID() uint64 {
	return this.id
}

func (this *Session) InterData() MsgParse {
	return this.MsgParse
}

func (this *Session) Send(data []byte) bool {
	msg := bufferPool.Alloc(len(data))
	msg = msg[:len(data)]
	copy(msg, data)
	for {
		select {
		case <-this.closer:
			return false
		case this.writer <- msg:
			return true
		case <-time.After(200 * time.Millisecond):
			break
		}
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
			bufferPool.Free(buf)
		}
	}

exitsend:
	close(this.closer)
}

func (this *Session) dorecv() {
	this.ProcMsg(this, SessionMsg{CMD_NEW, nil})

	msgbuf := bufferPool.Alloc(MsgBuffSize)
	defer bufferPool.Free(msgbuf)
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
			msgcpy := bufferPool.Alloc(len(msg))
			msgcpy = msgcpy[:len(msg)]
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
