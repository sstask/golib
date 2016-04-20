package stnet

import (
	"net"
	"time"
)

type Connector struct {
	sess          *Session
	address       string
	reconnectMSec int //Millisecond
	closeSignal   chan int
	exitSignal    chan int
}

func NewConnector(address string, reconnectmsec int, msgparse MsgParse) *Connector {
	if msgparse == nil {
		return nil
	}

	conn := &Connector{
		closeSignal:   make(chan int),
		exitSignal:    make(chan int),
		address:       address,
		reconnectMSec: reconnectmsec,
	}

	go func() {
		for {
			cn, err := net.Dial("tcp", conn.address)
			if err != nil {
				if conn.reconnectMSec == 0 {
					break
				}
				time.Sleep(time.Duration(conn.reconnectMSec) * time.Millisecond)
				continue
			}

			conn.sess = NewSession(cn, msgparse, func(*Session) {
				conn.closeSignal <- 1
			})

			_, ok := <-conn.closeSignal
			if !ok {
				break //chan closed
			}
			if conn.reconnectMSec == 0 {
				break
			}
			time.Sleep(time.Duration(conn.reconnectMSec) * time.Millisecond)
		}
		conn.exitSignal <- 1
	}()
	return conn
}

func (this *Connector) Send(data []byte) bool {
	if this.sess == nil {
		return false
	}
	return this.sess.Send(data)
}

func (this *Connector) IsConnected() bool {
	if this.sess == nil {
		return false
	}
	return !this.sess.IsClose()
}

func (this *Connector) Close() {
	this.reconnectMSec = 0
	if this.sess != nil {
		this.sess.Close()
	}
	<-this.exitSignal
	close(this.exitSignal)
}
