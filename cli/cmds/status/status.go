package status

import (
	svc "github.com/kardianos/service"
	"github.com/zhiyunliu/velocity/cli/cmds"
	"github.com/zhiyunliu/velocity/cli/cmds/service"
	"github.com/zhiyunliu/velocity/configs"

	"github.com/urfave/cli"
)

func init() {
	cmds.RegisterFunc(func(cfg *configs.AppSetting) cli.Command {
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

var statusMap = map[svc.Status]string{
	svc.StatusRunning: "Running",
	svc.StatusStopped: "Stopped",
	svc.StatusUnknown: "Unknown",
}
