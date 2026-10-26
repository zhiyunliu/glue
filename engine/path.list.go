package engine

type RouterList struct {
	ServerType string      `json:"server_type"  yaml:"server_type"`
	PathResult *PathResult `json:"path_result"  yaml:"path_result"`
}

func (r RouterList) GetType() string {
	return r.ServerType
}

func (r RouterList) GetPathResult() any {
	return r.PathResult
}
