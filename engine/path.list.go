package engine

type RouterList struct {
	ServerType string   `json:"server_type"  yaml:"server_type"`
	PathList   []string `json:"path_list"  yaml:"path_list"`
}

func (r RouterList) GetType() string {
	return r.ServerType
}

func (r RouterList) GetPathList() []string {
	return r.PathList
}
