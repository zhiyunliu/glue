package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&compareNormalMatcher{
		//> aaa
		//> t.aaa
		//>= aaa
		//>= t.aaa

		//< aaa
		//< t.aaa
		//<= aaa
		//<= t.aaa

		regexp: regexp.MustCompile(`^(>|>=|<|<=)\s*(\w+(\.\w+)?)$`),
	})
}

type compareNormalMatcher struct {
	regexp *regexp.Regexp
}

func (m *compareNormalMatcher) Priority() int {
	return 20
}

func (m *compareNormalMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *compareNormalMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(fullkey)
	if len(parties) <= 0 {
		return
	}
	ok = true
	var (
		item = &PropItem{}
	)
	item.Oper = parties[1]
	fullField := parties[2]

	fullField = strings.TrimSpace(fullField)
	item.FullField = fullField
	item.PropName = fullField

	idx := strings.Index(fullField, ".")
	if idx > 0 {
		item.PropName = fullField[idx+1:]
	}

	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *compareNormalMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
