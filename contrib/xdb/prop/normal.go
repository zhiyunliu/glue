package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&normalMatcher{
		regexp: regexp.MustCompile(`^(\w+(\.\w+)?)$`),
	})
}

type normalMatcher struct {
	regexp *regexp.Regexp
}

func (m *normalMatcher) Priority() int {
	return 0
}

func (m *normalMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *normalMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {
	ok = m.regexp.MatchString(fullkey)
	if !ok {
		return
	}

	parties := strings.Split(fullkey, ".")

	item := &PropItem{
		Oper:      "=",
		FullField: fullkey,
		PropName:  fullkey,
	}
	if len(parties) > 1 {
		item.PropName = parties[1]
	}

	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *normalMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {

		




		return
	}
}
