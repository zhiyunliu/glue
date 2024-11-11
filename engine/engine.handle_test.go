package engine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/golibs/xtypes"
)

type mockLogger struct {
	output string
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.output = fmt.Sprintf(format, args...)
}

func TestPrintSource_HappyPath(t *testing.T) {
	logOpts := &log.Options{WithSource: &[]bool{true}[0]}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: &[]bool{true}[0]}}
	header := xtypes.SMap{
		constants.HeaderSourceIp:   "127.0.0.1",
		constants.HeaderSourceName: "testApp",
	}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)

	assert.Contains(t, logger.output, "X-Src-Ip=127.0.0.1")
	assert.Contains(t, logger.output, "X-Src-Name=testApp")

}

func TestPrintSource_EmptyHeader(t *testing.T) {
	logOpts := &log.Options{WithSource: &[]bool{true}[0]}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: &[]bool{true}[0]}}
	header := xtypes.SMap{}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)

	assert.Empty(t, logger.output)
}

func TestPrintSource_NoPrintSource(t *testing.T) {
	logOpts := &log.Options{WithSource: &[]bool{false}[0]}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: &[]bool{false}[0]}}
	header := xtypes.SMap{
		constants.HeaderSourceIp:   "127.0.0.1",
		constants.HeaderSourceName: "testApp",
	}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)

	assert.Empty(t, logger.output)
}

func TestPrintSource_NoPrintSourceNotSet(t *testing.T) {
	logOpts := &log.Options{WithSource: &[]bool{true}[0]}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: nil}}
	header := xtypes.SMap{
		constants.HeaderSourceIp:   "127.0.0.1",
		constants.HeaderSourceName: "testApp",
	}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)

	assert.Contains(t, logger.output, "X-Src-Ip=127.0.0.1")
	assert.Contains(t, logger.output, "X-Src-Name=testApp")
}

func TestPrintSource_NoPrintSourcePath(t *testing.T) {
	logOpts := &log.Options{WithSource: &[]bool{true}[0]}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: &[]bool{false}[0]}}
	header := xtypes.SMap{
		constants.HeaderSourceIp:   "127.0.0.1",
		constants.HeaderSourceName: "testApp",
	}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)

	assert.Empty(t, logger.output)
}

func TestPrintSource_NilWithSource(t *testing.T) {
	logOpts := &log.Options{WithSource: nil}
	group := &RouterWrapper{opts: &RouterOptions{WithSource: nil}}
	header := xtypes.SMap{
		constants.HeaderSourceIp:   "127.0.0.1",
		constants.HeaderSourceName: "testApp",
	}
	logger := &mockLogger{}
	printSource(logger, logOpts, group, header)
}
