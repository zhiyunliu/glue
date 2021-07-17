package status

import (
	"github.com/lib4dev/cli/cmds"

	"github.com/urfave/cli"
)

func init() {
	cmds.RegisterFunc(func() cli.Command {
		return cli.Command{
			Name:   "status",
			Usage:  "查询状态，查询服务器运行、停止状态",
			Action: doStatus,
		}
	})
}

func doStatus(c *cli.Context) (err error) {

	//3.创建本地服务
	velocitySrv, err := service.GetService(c)
	if err != nil {
		return err
	}
	status, err := velocitySrv.Status()
	return service.GetCmdsResult(velocitySrv.DisplayName, "Status", err, statusMap[status])
}

var statusMap = map[service.Status]string{
	service.StatusRunning: "Running",
	service.StatusStopped: "Stopped",
	service.StatusUnknown: "Unknown",
}
