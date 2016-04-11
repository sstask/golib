package stnet

import (
	"net"
	"sync/atomic"
)

const (
	CMD_NEW int = iota
	CMD_CLOSE
	CMD_DATA
)

type SessionMsg struct {
	Cmd  int
	Data []byte
}
type FuncProcMsg func(*Session, SessionMsg)
type FuncParseMsg func(buf []byte) (parsedlen int, msg []byte) //buf:recved data now;parsedlen:length of recved data parsed;msg: message which is parsed from recved data
type FuncOnClose func(*Session)                                //close event

const MsgBuffSize = 1024
const WriterListLen = 256

var GlobalSessionID uint64

type Session struct {
	id     uint64
	socket net.Conn
	writer chan []byte
	closer chan int
	wclose chan int

	procmsg  FuncProcMsg
	parsemsg FuncParseMsg
	onclose  FuncOnClose
}

func NewSession(con net.Conn, parsemsg FuncParseMsg, procmsg FuncProcMsg, onclose FuncOnClose) *Session {
	if parsemsg == nil || procmsg == nil {
		return nil
	}
	sess := &Session{
		id:       atomic.AddUint64(&GlobalSessionID, 1),
		socket:   con,
		writer:   make(chan []byte, WriterListLen),
		closer:   make(chan int),
		wclose:   make(chan int),
		parsemsg: parsemsg,
		procmsg:  procmsg,
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
	this.procmsg(this, SessionMsg{CMD_NEW, nil})

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
		dellen, msg := this.parsemsg(msgbuf[0:msglen])
		if msg != nil {
			msgcpy := make([]byte, len(msg))
			copy(msgcpy, msg)
			this.procmsg(this, SessionMsg{CMD_DATA, msgcpy})
		}
		if dellen > 0 && dellen <= msglen {
			if dellen < msglen {
				copy(msgbuf, msgbuf[dellen:msglen])
			}
			msglen -= dellen
		}
	}

exitrecv:
	this.procmsg(this, SessionMsg{CMD_CLOSE, nil})
	close(this.wclose)
	this.socket.Close()
	<-this.closer //wait send routine exit
	if this.onclose != nil {
		this.onclose(this)
	}
}
