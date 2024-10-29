package prop

import (
	"regexp"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&compareSpecMatcher{
		//t.field > aaa
		//t.field > t.aaa
		//t.field >= aaa
		//t.field >= t.aaa

		//field < aaa
		//field < aaa
		//field <= aaa
		//field <= aaa

		regexp: regexp.MustCompile(`^((\w+\.)?\w+)\s*(>|>=|<|<=)\s*(\w+)$`),
	})
}

type compareSpecMatcher struct {
	regexp *regexp.Regexp
}

func (m *compareSpecMatcher) Priority() int {
	return 21
}

func (m *compareSpecMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *compareSpecMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(fullkey)
	if len(parties) <= 0 {
		return
	}
	ok = true

	item := &PropItem{
		FullField: parties[1],
		Oper:      parties[2],
		PropName:  parties[3],
	}
	if len(parties) == 5 {
		item.Oper = parties[3]
		item.PropName = parties[4]
	}

	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *compareSpecMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
