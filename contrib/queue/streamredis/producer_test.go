package streamredis

import (
	"testing"
)

func tmpFunc(buildLen int) (procCnt int) {
	if buildLen == 0 {
		return
	}
	vals := make([]string, buildLen)

	const CMD_COUNT = 100
	tmpLen := len(vals)
	if tmpLen > CMD_COUNT {
		tmpLen = CMD_COUNT
	}

	cycCnt := len(vals) / tmpLen
	if cycCnt*tmpLen < len(vals) {
		cycCnt = cycCnt + 1
	}

	idx := 0
	totalLen := len(vals)
	isLast := false

	for c := 0; c < cycCnt; c++ {
		args := make([]interface{}, 0, tmpLen)
		cycIdx := c * tmpLen
		isLast = c == (cycCnt - 1)

		for i := 0; i < tmpLen; i++ {
			idx = cycIdx + i
			args = append(args, vals[idx])
			if isLast && (idx+1) == totalLen {
				break
			}
		}
		procCnt += len(args)
		// err = p.client.ZRem(p.opts.DelayQueueName, args...).Err()
		// if err != nil {
		// 	log.Errorf("streamredis.procDelayQueue.ZRem:%s,err:%+v", p.opts.DelayQueueName, err)
		// }
	}
	return procCnt
}

func TestProducer_procDelayQueue(t *testing.T) {
	tests := []struct {
		name string
		cur  int64
	}{
		{name: "1.", cur: 0},
		{name: "1.", cur: 1},
		{name: "2.", cur: 100},
		{name: "3.", cur: 101},
		{name: "4.", cur: 200},
		{name: "5.", cur: 201},
		{name: "6.", cur: 310},
		{name: "7.", cur: 300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tmpFunc(int(tt.cur))
			if r != int(tt.cur) {
				t.Errorf("%s cur:%d,r:%d ", tt.name, tt.cur, r)
			}
		})
	}
}
