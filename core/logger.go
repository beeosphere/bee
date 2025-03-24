package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Logger is an interface that defines the logging methods.
type Logger interface {
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// logrusLogger is a wrapper around logrus.Logger that implements the Logger interface.
type logrusLogger struct {
	logger  *logrus.Logger
	agent   string
	channel string
	prefix  string
}

func (l *logrusLogger) SetPrefix(agent, channel string) {
	l.agent = agent
	l.channel = channel
	if channel == "" {
		l.prefix = fmt.Sprintf("[%s] ", agent)
	} else {
		l.prefix = fmt.Sprintf("[%s/%s] ", agent, channel)
	}
}

func (l *logrusLogger) SetLevel(level logrus.Level) {
	l.logger.SetLevel(level)
}

// newLogrusLogger creates a new LogrusLogger.
func newLogrusLogger() *logrusLogger {
	return &logrusLogger{
		logger: logrus.StandardLogger(),
	}
}

// Debug logs a debug message.
func (l *logrusLogger) Trace(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Trace(args...)
}

// Debugf logs a debug message using a format string.
func (l *logrusLogger) Tracef(format string, args ...interface{}) {
	l.logger.Tracef(l.prefix+format, args...)
}

// Debug logs a debug message.
func (l *logrusLogger) Debug(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Debug(args...)
}

// Debugf logs a debug message using a format string.
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(l.prefix+format, args...)
}

// Info logs an info message.
func (l *logrusLogger) Info(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Info(args...)
}

// Infof logs an info message using a format string.
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(l.prefix+format, args...)
}

// Warn logs a warning message.
func (l *logrusLogger) Warn(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(l.prefix+format, args...)
}

// Error logs an error message.
func (l *logrusLogger) Error(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Error(args...)
}

// Errorf logs an error message using a format string.
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(l.prefix+format, args...)
}
