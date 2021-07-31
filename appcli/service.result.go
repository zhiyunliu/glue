package appcli

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/velocity/compatible"
)

//buildCmdResult  buildCmdResult
func buildCmdResult(serviceName, action string, err error, args ...string) error {
	if err != nil {
		return fmt.Errorf("%s %s %s:%w", action, serviceName, compatible.FAILED, err)
	}
	if len(args) > 0 {
		serviceName = serviceName + " " + strings.Join(args, " ")
	}
	return fmt.Errorf("%s %s %s", action, serviceName, compatible.SUCCESS)
}
