package service

import (
	"github.com/fuguohong1024/sentinel-proxy/internal/app/core"
	"io"
	"net"
)

type ProxyBridge struct{}

func (pb *ProxyBridge) Proxy(clientConn, redisConn net.Conn) {
	go pb.proxyConnection(clientConn, redisConn)
	go pb.proxyConnection(redisConn, clientConn)
}

func (pb *ProxyBridge) proxyConnection(destConn net.Conn, srcConn net.Conn) {
	_, err := io.Copy(destConn, srcConn)
	if err != nil && err != io.EOF {
		core.GetLogger().Debug("Proxy track closed")
	}

	_ = destConn.Close()
	_ = srcConn.Close()
}
