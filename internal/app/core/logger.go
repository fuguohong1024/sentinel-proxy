package core

import (
	"fmt"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	"github.com/sirupsen/logrus"
	"os"
)

type Logger struct {
	*logrus.Logger
}

var logger *Logger

func GetLogger() *Logger {
	if logger == nil {
		logger = initLogger()
	}

	return logger
}

func initLogger() *Logger {
	config := GetConfig()
	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logLevel = logrus.DebugLevel
	}

	externalLogger := logrus.New()
	externalLogger.Out = os.Stdout
	externalLogger.Level = logLevel
	externalLogger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		// 2006-01-02 15:04:05
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
	})

	if len(config.GraylogHost) > 0 && len(config.GraylogPort) > 0 {
		graylogCredentials := fmt.Sprintf("%s:%s", config.GraylogHost, config.GraylogPort)
		hook := graylog.NewGraylogHook(graylogCredentials, map[string]interface{}{"service": "sentinel_proxy"})
		externalLogger.AddHook(hook)
	}

	return &Logger{externalLogger}
}
