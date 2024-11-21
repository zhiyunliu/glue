package cli

import (
	svc "github.com/kardianos/service"

	"github.com/urfave/cli"
)

func init() {
	RegisterFunc(func(cfg *Options) *cli.Command {
		return &cli.Command{
			Name:   "status",
			Usage:  "查询状态，查询服务器运行、停止状态",
			Flags:  getFlags(cfg),
			Action: doStatus,
		}
	})
}

func doStatus(c *cli.Context) (err error) {

	//创建本地服务
	velocitySrv, err := getService(c)
	if err != nil {
		return err
	}
	status, err := velocitySrv.Status()
	return buildCmdResult(velocitySrv.DisplayName, "Status", err, statusMap[status])
}

var statusMap = map[svc.Status]string{
	svc.StatusRunning: "Running",
	svc.StatusStopped: "Stopped",
	svc.StatusUnknown: "Unknown",
}
