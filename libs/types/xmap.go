package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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

//Get 获取指定元素的Bool值
func (m XMap) GetBool(name string) bool {
	v, ok := m[name]
	if ok {
		tmp, err := strconv.ParseBool(fmt.Sprint(v))
		if err != nil {
			return false
		}
		return tmp
	}
	return false
}

//Get 获取指定元素的值
func (m XMap) GetString(name string) string {
	v, ok := m[name]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%+v", v)
}

func (m XMap) Scan(obj interface{}) error {
	bytes, _ := json.Marshal(m)
	return json.Unmarshal(bytes, obj)
}

func (m XMap) Len() int {
	return len(m)
}

func (m XMap) SMap() SMap {
	sm := map[string]string{}
	for k, v := range m {
		sm[k] = fmt.Sprint(v)
	}
	return sm
}
