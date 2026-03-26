package core

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// LogrusLogger is a wrapper around logrus.Logger that implements the Logger interface.
type LogrusLogger struct {
	logger  *logrus.Logger
	agent   string
	channel string
	prefix  string
}

func (l *LogrusLogger) SetPrefix(agent, channel string) {
	l.agent = agent
	l.channel = channel
	if channel == "" {
		l.prefix = fmt.Sprintf("[%s] ", agent)
	} else {
		l.prefix = fmt.Sprintf("[%s/%s] ", agent, channel)
	}
}

func (l *LogrusLogger) SetLevel(level logrus.Level) {
	l.logger.SetLevel(level)
}

// newLogrusLogger creates a new LogrusLogger.
func NewLogrusLogger() *LogrusLogger {
	return &LogrusLogger{
		logger: logrus.StandardLogger(),
	}
}

// Debug logs a debug message.
func (l *LogrusLogger) Trace(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Trace(args...)
}

// Debugf logs a debug message using a format string.
func (l *LogrusLogger) Tracef(format string, args ...interface{}) {
	l.logger.Tracef(l.prefix+format, args...)
}

// Debug logs a debug message.
func (l *LogrusLogger) Debug(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Debug(args...)
}

// Debugf logs a debug message using a format string.
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(l.prefix+format, args...)
}

// Info logs an info message.
func (l *LogrusLogger) Info(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Info(args...)
}

// Infof logs an info message using a format string.
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(l.prefix+format, args...)
}

// Warn logs a warning message.
func (l *LogrusLogger) Warn(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Warn(args...)
}

func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(l.prefix+format, args...)
}

// Error logs an error message.
func (l *LogrusLogger) Error(args ...interface{}) {
	args = append([]interface{}{l.prefix}, args...)
	l.logger.Error(args...)
}

// Errorf logs an error message using a format string.
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(l.prefix+format, args...)
}
