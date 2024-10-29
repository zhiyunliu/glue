package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&inSpecMatcher{
		//t.aaa in aaa
		//aaa in aaa
		regexp: regexp.MustCompile(`^(\w+(\.\w+)?)\s+in\s+(\w+)$`),
	})
}

type inSpecMatcher struct {
	regexp *regexp.Regexp
}

func (m *inSpecMatcher) Priority() int {
	return 31
}

func (m *inSpecMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *inSpecMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(fullkey)
	if len(parties) <= 0 {
		return
	}

	var (
		item = &PropItem{
			Oper: "in",
		}
	)
	fullField := parties[1]

	fullField = strings.TrimSpace(fullField)
	item.FullField = fullField

	item.PropName = parties[3]

	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *inSpecMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
