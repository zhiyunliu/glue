package logger

import (
	"os"

	"github.com/rs/zerolog"
)

type zeroLogger struct{}

func (l *zeroLogger) Debug(content ...interface{})                 {}
func (l *zeroLogger) Debugf(format string, content ...interface{}) {}
func (l *zeroLogger) Info(content ...interface{})                  {}
func (l *zeroLogger) Infof(format string, content ...interface{})  {}
func (l *zeroLogger) Error(content ...interface{})                 {}
func (l *zeroLogger) Errorf(format string, content ...interface{}) {}
func (l *zeroLogger) Fatal(content ...interface{})                 {}
func (l *zeroLogger) Fatalf(format string, content ...interface{}) {}
func (l *zeroLogger) Close()
func (l *zeroLogger) SetSessionID(sid string) {

}

func newZeroLogger() *zeroLogger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)

	logger := zerolog.New(multi).With().Timestamp().Logger()

	logger.Info().Msg("Hello World!")
}
