package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func Init() {
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(logrus.InfoLevel)
	Logger.Info("Structured logger initialized")
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}
