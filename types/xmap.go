package types

type XMap map[string]interface{}

//Keys 从对象中获取数据值，如果不是字符串则返回空
func (m XMap) Keys() []string {
	keys := make([]string, len(m))
	idx := 0
	for k := range m {
		keys[idx] = k
		idx++
	}
	return keys
}

//Merge 合并
func (m XMap) Merge(r XMap) {
	for k, v := range r {
		m[k] = v
	}
}

//Get 获取指定元素的值
func (m XMap) Get(name string) (interface{}, bool) {
	v, ok := m[name]
	return v, ok
}

func (m XMap) Len() int {
	return len(m)
}
