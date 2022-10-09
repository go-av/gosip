package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.AddHook(NewVideoHook())
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

type Level int

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

func SetLevel(level Level) {
	logrus.SetLevel(logrus.Level(level))
}

func NewVideoHook() *VideoHook {
	return &VideoHook{}
}

type VideoHook struct {
}

func (hook *VideoHook) Fire(entry *logrus.Entry) error {
	entry.Data["type"] = "gosip"
	return nil
}

func (hook *VideoHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
