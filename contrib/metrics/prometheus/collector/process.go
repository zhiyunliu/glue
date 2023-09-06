package collector

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	processCPUPercent = prometheus.NewDesc(
		"namedprocess_cpu_percent",
		"named process cpu percentage",
		[]string{"processname"},
		nil)
)

type (
	NamedProcessCollector struct {
		processInfo *process.Process
		name        string
	}
)

func NewProcessCollector() (p *NamedProcessCollector, err error) {

	processes, err := process.Processes()
	if err != nil {
		return
	}
	curPid := os.Getpid()
	if err != nil {
		return
	}
	var curProcess *process.Process
	for _, p := range processes {
		if p.Pid == int32(curPid) {
			curProcess = p
			break
		}
	}
	name, err := curProcess.Name()
	if err != nil {
		return
	}
	p = &NamedProcessCollector{
		name:        name,
		processInfo: curProcess,
	}

	return p, nil
}

// Describe implements prometheus.Collector.
func (p *NamedProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- processCPUPercent
}

// Collect implements prometheus.Collector.
func (p *NamedProcessCollector) Collect(ch chan<- prometheus.Metric) {
	cpuPercent, err := p.processInfo.CPUPercent()
	if err != nil {
		cpuPercent = -1
	}
	ch <- prometheus.MustNewConstMetric(processCPUPercent, prometheus.CounterValue, cpuPercent, p.name)
}
