package xtypes

type XMaps []XMap

func (ms *XMaps) Append(i ...XMap) XMaps {
	*ms = append(*ms, i...)
	return *ms
}

func (ms XMaps) IsEmpty() bool {
	return ms == nil || len(ms) == 0
}

func (ms XMaps) Len() int {
	return len(ms)
}
