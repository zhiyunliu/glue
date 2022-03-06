package config

type Setting struct {
	rawBytes []byte
}

//todo:https://github.com/oliveagle/jsonpath 是否考虑使用
func (s *Setting) Get(name string) *Setting {
	return s
	// json.

	// if s.data == nil {
	// 	return nil
	// }
	// val, ok := s.data[name]
	// if !ok {
	// 	return nil
	// }
	// return val.(*Setting)
}

func (s *Setting) GetRaw() []byte {
	return s.rawBytes
}

func (s *Setting) GetProperty(name string) string {
	return ""
}

func (s *Setting) Scan(obj interface{}) error {
	return nil
}
