package log

import (
	"github.com/Sirupsen/logrus"
)

var defaultLogger *logrus.Logger

func DefaultLogger() *logrus.Logger {
	if defaultLogger == nil {
		defaultLogger = logrus.New()
	}
	return defaultLogger
}
