package service

import (
	"net"
	"sync"
)

type track struct {
	upStream   net.Conn
	downStream net.Conn
}

func (t *track) SetKeepAlive(keepAive bool) {
	t.upStream.(*net.TCPConn).SetKeepAlive(keepAive)
	t.downStream.(*net.TCPConn).SetKeepAlive(keepAive)
}

// SetNoDelay tcp noDelay 是否合并小包
func (t *track) SetNoDelay(noDelay bool) {
	t.upStream.(*net.TCPConn).SetNoDelay(noDelay)
	t.downStream.(*net.TCPConn).SetKeepAlive(noDelay)
}

type Proxy struct {
	locker *sync.Mutex
	Tracks map[string]track
}

func NewProxy() *Proxy {
	locker := new(sync.Mutex)
	return &Proxy{locker: locker,
		Tracks: make(map[string]track),
	}
}

func (p *Proxy) AddConnection(c string, upStream, downStram net.Conn) {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.Tracks[c] = track{
		upStream:   upStream,
		downStream: downStram,
	}
}

func (p *Proxy) DeleteConnection(c string) {
	p.locker.Lock()
	defer p.locker.Unlock()

	delete(p.Tracks, c)
}

func (p *Proxy) Close() {
	p.locker.Lock()
	defer p.locker.Unlock()
	for _, tra := range p.Tracks {
		tra.upStream.Close()
		tra.downStream.Close()
	}
}
