package gfwlist

import (
	"regexp"
	"strings"
	"sync"

	"github.com/denghongcai/yaproxy/cache"
	"github.com/denghongcai/yaproxy/util"
)

var Parser = GFWListParser{}

type GFWListParser struct {
	sync.RWMutex
	rules []*regexp.Regexp
}

func (this *GFWListParser) LoadGFWList(code string) {
	defer this.Unlock()
	this.Lock()
	rawRules := strings.Split(code, "\n")
	for _, rawRule := range rawRules {
		rawRule = strings.Trim(rawRule, " \r\n")
		if strings.HasPrefix(rawRule, "!") {
			continue
		}
		if strings.HasPrefix(rawRule, "/") && strings.HasSuffix(rawRule, "/") {
			this.rules = append(this.rules, regexp.MustCompile(rawRule[1:len(rawRule)-1]))
		}
		if strings.HasPrefix(rawRule, "||") {
			this.rules = append(this.rules, regexp.MustCompile(`.+://`+rawRule[2:]))
		}
		if strings.HasPrefix(rawRule, ".") {
			this.rules = append(this.rules, regexp.MustCompile(`.+`+rawRule+`.+`))
		}
	}
}

func (this *GFWListParser) NeedProxy(host string, port int) bool {
	url := util.BuildURL(host, port)
	if value, exists := cache.TestURL(url); exists {
		return value
	} else {
		for _, rule := range this.rules {
			if rule.MatchString(url) {
				return true
			}
		}
		return false
	}
}
