package logger

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"time"
)

func InitLogger(debug bool) *logrus.Logger {
	Log := logrus.New()

	var logLevel = logrus.InfoLevel
	if debug {
		logLevel = logrus.DebugLevel
	}

	Log.SetLevel(logLevel)
	Log.SetOutput(colorable.NewColorableStdout())
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})

	return Log
}
