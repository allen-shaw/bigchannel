package log

import "github.com/allen-shaw/log"

type Logger interface {
}

func NewLogger() Logger {

	return newLogger()
}

type logger struct {
	l *log.Logger
}

func newLogger() *logger {
	l := &logger{}
	return l
}
