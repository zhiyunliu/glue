package xdb

type DefaultSqlStateCahe struct {
	tplSql        string
	names         []string
	nameCache     map[string]string
	hasReplace    bool
	hasDynamicAnd bool
	hasDynamicOr  bool
	ph            Placeholder
	SQLTemplate   SQLTemplate
}

func (item *DefaultSqlStateCahe) ClonePlaceHolder() Placeholder {
	return item.ph.Clone()
}

func (item *DefaultSqlStateCahe) Build(input DBParam) (execSql string, values []interface{}, err error) {
	values = make([]interface{}, len(item.names))
	ph := item.ClonePlaceHolder()
	var outerrs []MissError
	var ierr MissError
	for i := range item.names {
		_, values[i], ierr = input.Get(item.names[i], ph)
		if ierr != nil {
			outerrs = append(outerrs, ierr)
		}
	}
	if len(outerrs) > 0 {
		return "", values, NewMissListError(outerrs...)
	}

	return execSql, values, err
}
