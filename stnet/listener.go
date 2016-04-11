package stnet

import (
	"net"
	"sync"
)

type Listener struct {
	isclose bool
	address string
	lst     net.Listener

	sessMap      map[uint64]*Session
	sessMapMutex sync.RWMutex
	waitExit     sync.WaitGroup
}

func NewListener(address string, parsemsg FuncParseMsg, procmsg FuncProcMsg) *Listener {
	if parsemsg == nil || procmsg == nil {
		return nil
	}

	ls, err := net.Listen("tcp", address)
	if err != nil {
		return nil
	}

	lis := &Listener{
		isclose: false,
		address: address,
		lst:     ls,
	}

	go func() {
		for {
			conn, err := lis.lst.Accept()
			if err != nil {
				break
			}

			lis.sessMapMutex.Lock()
			lis.waitExit.Add(1)
			sess := NewSession(conn, parsemsg, procmsg, func(con *Session) {
				lis.sessMapMutex.Lock()
				delete(lis.sessMap, con.id)
				lis.waitExit.Done()
				lis.sessMapMutex.Unlock()
			})
			lis.sessMap[sess.id] = sess
			lis.sessMapMutex.Unlock()
		}
		lis.Close()
	}()
	return lis
}

func (this *Listener) Close() {
	if this.isclose {
		return
	}
	this.isclose = true
	this.lst.Close()
	this.IterateSession(func(sess *Session) bool {
		sess.Close()
		return true
	})
	this.waitExit.Wait()
}

func (this *Listener) GetSession(id uint64) *Session {
	this.sessMapMutex.RLock()
	defer this.sessMapMutex.RUnlock()

	v, ok := this.sessMap[id]
	if ok {
		return v
	}
	return nil
}

func (this *Listener) IterateSession(callback func(*Session) bool) {
	this.sessMapMutex.RLock()
	defer this.sessMapMutex.RUnlock()

	for _, ses := range this.sessMap {
		if !callback(ses) {
			break
		}
	}
}
