package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	if os.Getenv("DEBUG") != "" {
		Logger.SetLevel(logrus.DebugLevel)
	}
}

func Log() *logrus.Logger {
	return Logger
}
