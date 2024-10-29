package prop

import (
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	RegistMatcher(&likeNormalMatcher{
		regexp: regexp.MustCompile(`^(like\s+%?\w+(\.\w+)?%?)$`),
	})
}

type likeNormalMatcher struct {
	regexp *regexp.Regexp
}

func (m *likeNormalMatcher) Priority() int {
	return 10
}

func (m *likeNormalMatcher) Pattern() string {
	return m.regexp.String()
}

func (m *likeNormalMatcher) MatchString(fullkey string) (valuer xdb.PropValuer, ok bool) {
	ok = m.regexp.MatchString(fullkey)
	if !ok {
		return
	}
	fullkey = strings.TrimPrefix(fullkey, "like")
	fullkey = strings.TrimSpace(fullkey)

	var (
		prefix string
		suffix string
		oper   string = "like"
	)

	if strings.HasPrefix(fullkey, "%") {
		prefix = "%"
	}
	if strings.HasSuffix(fullkey, "%") {
		suffix = "%"
	}

	oper = prefix + oper + suffix
	fullkey = strings.Trim(fullkey, "%")

	item := &PropItem{
		Oper:      oper,
		FullField: fullkey,
	}
	item.PropName = getPropertyName(fullkey)
	item.PropertyCallback = m.buildCallback()
	return item, ok
}

func (m *likeNormalMatcher) buildCallback() PropertyCallback {
	return func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
		return
	}
}
