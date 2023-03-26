package frame2

import (
	"log"
)

type FrameLogger interface {
	// passes its parameters to a log.Logger's Printf.  Either
	// the one returned by GetLogger, or the default logger if
	// that call returns nil
	Printf(format string, v ...any)
	GetLogger() *log.Logger
	SetLogger(logger *log.Logger)

	// Sets the logger, but only if it is currently nil
	OrSetLogger(logger *log.Logger)
}

// Unconfigured, this simply forwards any Printf calls to the
// default logger.
//
// If a log.Logger is set, use it instead.
type Log struct {
	Logger *log.Logger
}

func (l Log) Printf(format string, v ...any) {
	if l.Logger == nil {
		log.Printf(format, v...)
	} else {
		l.Logger.Printf(format, v...)
	}
}

func (l *Log) GetLogger() *log.Logger {
	if l == nil {
		return nil
	}
	return l.Logger
}

func (l *Log) SetLogger(logger *log.Logger) {
	l.Logger = logger
}

// Sets the logger, but only if it is currently nil
// and the new value is not nil
func (l *Log) OrSetLogger(logger *log.Logger) {
	if l == nil {
		return
	}
	if l.Logger == nil {
		(*l).Logger = logger
	}
}

// This checks that its input is a frame2.Logger, then calls
// OrSetLogger on it.
//
// Its only objective is to keep the code concise (ie, remove
// all those type checks from the main code)
//
// x should be a pointer to a FrameLogger instance
func OrSetLogger(x FrameLogger, logger *log.Logger) {
	if x == nil {
		return
	}
	if x, ok := x.(FrameLogger); ok {
		(x).OrSetLogger(logger)
	}
}
