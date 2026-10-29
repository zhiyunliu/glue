package defaultipc

import "github.com/zhiyunliu/glue/metrics"

type Metrics struct {
	IpcProcessing metrics.Int64UpDownCounter `metric:"ipc_proc"  `
}
