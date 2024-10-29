package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&likeSpecMatcher{
		//aaaa like ttt
		//aaaa like %ttt
		//aaaa like ttt%
		//aaaa like %ttt%
		//tt.aaaa like bbb
		//tt.aaaa like %bbb
		//tt.aaaa like bbb%
		//tt.aaaa like %bbb%
		regexp: regexp.MustCompile(`^(\w+(\.\w+)?)\s+like\s+(%?\w+%?)$`),
	})
}

type likeSpecMatcher struct {
	regexp *regexp.Regexp
}

func (m *likeSpecMatcher) Priority() int {
	return 11
}

func (m *likeSpecMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *likeSpecMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {

	parties := m.regexp.FindStringSubmatch(fullkey)
	if len(parties) <= 0 {
		return
	}
	ok = true
	var (
		prefix string
		suffix string
		oper   string = "like"
		item          = &PropItem{}
	)
	item.FullField = parties[1]

	propertykey := parties[2]
	if len(parties) == 4 {
		propertykey = parties[3]
	}

	propertykey = strings.TrimSpace(propertykey)

	if strings.HasPrefix(propertykey, "%") {
		prefix = "%"
	}
	if strings.HasSuffix(propertykey, "%") {
		suffix = "%"
	}

	oper = prefix + oper + suffix
	propertykey = strings.Trim(propertykey, "%")

	item.Oper = oper
	item.PropName = propertykey
	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *likeSpecMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
