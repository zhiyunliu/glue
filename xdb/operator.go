package xdb

var (
	//新建一个符号处理
	NewOperator func(name string, callback OperatorCallback) Operator
)

// OperatorCallback 操作符回调函数
type OperatorCallback func(valuer ExpressionValuer, param DBParam, phName string, value any) string

// Operator 操作符处理接口
type Operator interface {
	Name() string
	Callback(valuer ExpressionValuer, param DBParam, phName string, value any) string
}

// OperatorMap 操作符映射接口
type OperatorMap interface {
	//Store(name string, callback OperatorCallback)
	Load(name string) (Operator, bool)
	Clone() OperatorMap
	Range(func(name string, callback Operator) bool)
}

type operatorMap struct {
	//syncMap *sync.Map
	syncMap map[string]Operator
}

// NewOperatorMap 创建操作符映射
func NewOperatorMap(operators ...Operator) OperatorMap {
	operMap := &operatorMap{
		//syncMap: &sync.Map{},
		syncMap: make(map[string]Operator),
	}
	for _, oper := range operators {
		//operMap.syncMap.Store(oper.Name(), oper)
		operMap.syncMap[oper.Name()] = oper
	}
	return operMap
}

func (m *operatorMap) Load(name string) (Operator, bool) {
	//callback, ok := m.syncMap.Load(name)
	callback, ok := m.syncMap[name]
	if !ok {
		return nil, ok
	}
	return callback, ok
}

func (m *operatorMap) Clone() OperatorMap {
	clone := &operatorMap{
		//syncMap: &sync.Map{},
		syncMap: make(map[string]Operator),
	}
	// m.syncMap.Range(func(key, value any) bool {
	// 	clone.syncMap.Store(key.(string), value.(OperatorCallback))
	// 	return true
	// })

	for key, value := range m.syncMap {
		clone.syncMap[key] = value
	}
	return clone
}

func (m *operatorMap) Range(f func(name string, operator Operator) bool) {
	// m.syncMap.Range(func(key, value any) bool {
	// 	return f(key.(string), value.(Operator))
	// })
	for key, value := range m.syncMap {
		if !f(key, value) {
			break
		}
	}
}
