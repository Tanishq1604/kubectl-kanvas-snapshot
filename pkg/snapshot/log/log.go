package log

import (
	"fmt"
	"io"

	"github.com/layer5io/meshkit/logger"
	"github.com/sirupsen/logrus"
)

// Logger interface defines methods for logging
type Logger interface {
	Info(msg string)
	Debug(msg string)
	Warn(msg string)
	Error(err error)
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// MeshkitLogger is a wrapper around meshkit logger
type MeshkitLogger struct {
	Log logger.Handler
}

// SetupLogger creates a new logger instance
func SetupLogger(appName string, debug bool, w io.Writer) Logger {
	// Create a new meshkit logger
	logLevel := logrus.InfoLevel
	if debug {
		logLevel = logrus.DebugLevel
	}

	log, err := logger.New(appName, logger.Options{
		Format:   logger.TerminalLogFormat,
		Output:   w,
		LogLevel: int(logLevel),
	})

	if err != nil {
		// If we can't create the meshkit logger, fall back to a simple implementation
		return &fallbackLogger{writer: w, debug: debug}
	}

	return &MeshkitLogger{Log: log}
}

// MeshkitLogger implementation

func (l *MeshkitLogger) Info(msg string) {
	l.Log.Info(msg)
}

func (l *MeshkitLogger) Debug(msg string) {
	l.Log.Debug(msg)
}

func (l *MeshkitLogger) Warn(msg string) {
	l.Log.Warnf("%s", msg) // Using Warnf as Warn expects an error
}

func (l *MeshkitLogger) Error(err error) {
	l.Log.Error(err)
}

func (l *MeshkitLogger) Infof(format string, args ...interface{}) {
	l.Log.Infof(format, args...)
}

func (l *MeshkitLogger) Debugf(format string, args ...interface{}) {
	l.Log.Debugf(format, args...)
}

func (l *MeshkitLogger) Warnf(format string, args ...interface{}) {
	l.Log.Warnf(format, args...)
}

func (l *MeshkitLogger) Errorf(format string, args ...interface{}) {
	l.Log.Warnf(format, args...) // Using Warnf as Errorf is not available in the interface
}

// Fallback logger in case meshkit logger fails to initialize
type fallbackLogger struct {
	writer io.Writer
	debug  bool
}

func (l *fallbackLogger) Info(msg string) {
	l.Infof("%s\n", msg)
}

func (l *fallbackLogger) Debug(msg string) {
	if l.debug {
		l.Debugf("%s\n", msg)
	}
}

func (l *fallbackLogger) Warn(msg string) {
	l.Warnf("%s\n", msg)
}

func (l *fallbackLogger) Error(err error) {
	l.Errorf("%s\n", err.Error())
}

func (l *fallbackLogger) Infof(format string, args ...interface{}) {
	fmt.Fprintf(l.writer, "[INFO] "+format, args...)
}

func (l *fallbackLogger) Debugf(format string, args ...interface{}) {
	if l.debug {
		fmt.Fprintf(l.writer, "[DEBUG] "+format, args...)
	}
}

func (l *fallbackLogger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(l.writer, "[WARN] "+format, args...)
}

func (l *fallbackLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(l.writer, "[ERROR] "+format, args...)
}
