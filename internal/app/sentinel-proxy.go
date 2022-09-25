package app

import (
	"errors"
	"github.com/fuguohong1024/sentinel-proxy/internal/app/core"
	"github.com/fuguohong1024/sentinel-proxy/internal/app/service"
	"net"
)

type SentinelProxy struct {
	Config *core.Config
	Logger *core.Logger
}

var sentinelProxyInstance *SentinelProxy
var SentinelProxyBootstrapErr = errors.New("sentinel proxy bootstrap error")

func NewSentinelProxy(config *core.Config, logger *core.Logger) *SentinelProxy {
	sentinelProxyInstance = &SentinelProxy{
		Config: config,
		Logger: logger,
	}

	return sentinelProxyInstance
}

func GetSentinelProxyInstance() *SentinelProxy {
	return sentinelProxyInstance
}

func (proxy *SentinelProxy) Close() {
	// not implemented yet
}

func (proxy *SentinelProxy) Start() error {
	err := proxy.bootstrap()
	if err != nil {
		return err
	}

	return proxy.serve()
}

func (proxy SentinelProxy) serve() error {
	sentinelConnector := service.NewSentinelConnector(proxy.Config.SentinelList, proxy.Config.RequestTimeout)
	redisConnector := service.NewRedisConnector(proxy.Config.RequestTimeout)
	dbConnector := service.DbConnector{}
	proxyBridge := service.ProxyBridge{}

	for _, db := range proxy.Config.DbList {
		dbListener, err := dbConnector.Listen(db.LocalPort)
		if err != nil {
			return err
		}

		go func(db core.Master) {
			proxy.Logger.Infof("local proxy for db started, name: %s, port: %d", db.MasterName, db.LocalPort)

			for {
				clientConn, err := dbListener.Accept()
				if err != nil {
					proxy.Logger.Errorf(
						"accept new client connection error, name: %s, port: %d, error: %s",
						db.MasterName,
						db.LocalPort,
						err,
					)

					continue
				}

				proxy.Logger.Debugf("accept new client connection, name: %s, port: %d", db.MasterName, db.LocalPort)

				redisAddr, err := sentinelConnector.GetActualRedisAddr(db.MasterName)
				if err != nil {
					proxy.Logger.Error("connect to sentinels error: ", err)
					_ = clientConn.Close()

					continue
				}

				redisConn, err := redisConnector.Connect(redisAddr)
				if err != nil {
					proxy.Logger.Error("connect to redis error: ", err)
					_ = clientConn.Close()

					continue
				}
				// 连接池保活？
				redisConn.(*net.TCPConn).SetKeepAlive(true)

				clientConn.(*net.TCPConn).SetKeepAlive(true)

				proxyBridge.Proxy(clientConn, redisConn)
			}
		}(db)
	}

	return nil
}

func (proxy *SentinelProxy) bootstrap() error {
	sentinelConnector := service.NewSentinelConnector(proxy.Config.SentinelList, proxy.Config.RequestTimeout)
	redisConnector := service.NewRedisConnector(proxy.Config.RequestTimeout)
	isOk := true

	for _, db := range proxy.Config.DbList {
		redisAddr, err := sentinelConnector.GetActualRedisAddr(db.MasterName)
		if err != nil {
			proxy.Logger.Error("connect to sentinels error: ", err)
			isOk = false

			continue
		}

		err = redisConnector.Ping(redisAddr)
		if err != nil {
			proxy.Logger.Errorf("ping master redis failed, addr: %s, error: %s", redisAddr, err)
			isOk = false
		}
	}

	if !isOk {
		return SentinelProxyBootstrapErr
	}

	proxy.Logger.Info("bootstrap redis proxy success")

	return nil
}
