package alloter

import "math"

//Counter 任务执行次数
type Counter struct {
	executed int
}

//Get 获取任务执行次数
func (m *Counter) Get() int {
	return m.executed
}

//Increase 累加执行次数
func (m *Counter) Increase() {
	if m.executed >= math.MaxInt32 {
		m.executed = 1
	} else {
		m.executed++
	}
}

//Round 轮数信息
type Round struct {
	round int
}

func (m *Round) CanProc() bool {
	if m.round > 0 {
		m.round--
		return false
	}
	return true
}

//Reduce 减少任务等待轮数
func (m *Round) Reduce() {
	m.round--
}

//Get 获取任务在几轮后执行
func (m *Round) Current() int {
	return m.round
}

//Update 更新任务下次执行的轮数
func (m *Round) Update(v int) {
	m.round = v
}
