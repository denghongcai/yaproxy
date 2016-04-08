package pac

import (
	"fmt"
	"sync"

	"github.com/denghongcai/yaproxy/cache"
	"github.com/robertkrimen/otto"
)

var Parser = PacParser{}

type PacParser struct {
	sync.RWMutex
	vm *otto.Otto
}

func (this *PacParser) LoadPac(code string) {
	defer this.Unlock()
	this.Lock()
	this.vm = otto.New()
	_, err := this.vm.Run(code)
	if err != nil {
		panic(err)
	}
}

func (this *PacParser) NeedProxy(host string, port int) bool {
	url := "tcp://" + host
	if port == 80 {
		url = "http://" + host
	} else if port == 443 {
		url = "https://" + host
	}
	params := fmt.Sprintf("FindProxyForURL(\"%s\", \"%s\")", url, host)
	if value, exists := cache.TestURL(url); exists {
		return value
	} else {
		this.Lock()
		defer this.Unlock()
		v, _ := this.vm.Run(params)
		result, _ := v.ToString()
		b := result != "DIRECT" && result != "undefined"
		cache.AddURL(url, b)
		return b
	}
}
