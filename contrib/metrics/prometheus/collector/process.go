package collector

import (
	"fmt"
	"os"

	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessCPUCollector struct {
	cpuUsage    *prometheus.GaugeVec
	proc        *process.Process // 使用 gopsutil 的进程对象
	processname string
}

func NewProcessCPUCollector() (*ProcessCPUCollector, error) {
	var pid int32 = int32(os.Getpid())
	// 初始化进程对象
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("process %d not found: %v", pid, err)
	}
	processname, err := p.Name()
	if err != nil {
		return nil, fmt.Errorf("process %d ,get name failed: %v", pid, err)
	}

	return &ProcessCPUCollector{
		cpuUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "process_cpu_usage_percent",
				Help: "CPU usage percentage for target process",
			},
			[]string{"processname"},
		),
		proc:        p,
		processname: processname,
	}, nil
}

func (c *ProcessCPUCollector) Describe(ch chan<- *prometheus.Desc) {
	c.cpuUsage.Describe(ch)
}

func (c *ProcessCPUCollector) Collect(ch chan<- prometheus.Metric) {
	// 获取 CPU 使用率
	percent, err := c.proc.CPUPercent()
	if err != nil {
		log.Printf("Failed to get CPU percent: %v", err)
		return
	}
	// 更新指标
	c.cpuUsage.WithLabelValues(
		c.processname,
	).Set(percent)

	c.cpuUsage.Collect(ch)
}
