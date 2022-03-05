package config

type Setting struct {
	data map[string]interface{}
}

//todo:https://github.com/oliveagle/jsonpath 是否考虑使用
func (s *Setting) Get(name string) *Setting {
	if s.data == nil {
		return nil
	}
	val, ok := s.data[name]
	if !ok {
		return nil
	}
	return val.(*Setting)
}
