package engine

type Status string

//type Header xtypes.SMap

const (
	StatusStart = "start"
	StatusStop  = "stop"
)

type RunStatus uint

const (
	Unstarted RunStatus = 1
	Pause     RunStatus = 2
	Running   RunStatus = 4
	Pending   RunStatus = 8
	Stoped    RunStatus = 16
)
