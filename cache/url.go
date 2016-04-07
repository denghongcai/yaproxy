package cache

import (
	"bufio"
	"fmt"
	"io"

	LRU "github.com/hashicorp/golang-lru"
)

var urlCache *LRU.Cache

func init() {
	urlCache, _ = LRU.New(32767)
}

func AddURL(url string, is bool) {
	urlCache.Add(url, is)
}

func TestURL(url string) (bool, bool) {
	if value, exists := urlCache.Get(url); exists {
		return value.(bool), exists
	} else {
		return false, exists
	}
}

func DumpToWriter(w io.Writer) {
	writer := bufio.NewWriter(w)
	keys := urlCache.Keys()
	for _, k := range keys {
		v, _ := TestURL(k.(string))
		writer.WriteString(fmt.Sprintf("%s:%t\n", k, v))
	}
	writer.Flush()
}

func RecoverFromReader(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var k string
		var v bool
		fmt.Sscanf(scanner.Text(), "%s:%t", k, v)
		AddURL(k, v)
	}
}
