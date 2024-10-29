package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&inNormalMatcher{
		//in aaa
		//in t.aaa
		regexp: regexp.MustCompile(`^(in\s+(\w+(\.\w+)?))$`),
	})
}

type inNormalMatcher struct {
	regexp *regexp.Regexp
}

func (m *inNormalMatcher) Priority() int {
	return 30
}

func (m *inNormalMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *inNormalMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(fullkey)
	if len(parties) <= 0 {
		return
	}
	ok = true
	var (
		item = &PropItem{
			Oper: "in",
		}
	)
	fullField := parties[2]

	fullField = strings.TrimSpace(fullField)
	item.FullField = fullField
	item.PropName = getPropertyName(fullField)

	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *inNormalMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
