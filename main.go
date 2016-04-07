package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/proxy"

	"github.com/denghongcai/yaproxy/cache"
	"github.com/denghongcai/yaproxy/pac"
	"github.com/denghongcai/yaproxy/socks5"
)

var socks5Addr string
var listenAddr string
var timeout int
var pacFile string

func main() {
	flag.StringVar(&socks5Addr, "s", "127.0.0.1:1080", "specify socks5 server to use, default to 127.0.0.1:1080")
	flag.StringVar(&listenAddr, "l", "127.0.0.1:10800", "specify listen address, default to 127.0.0.1:10800")
	flag.IntVar(&timeout, "t", 2, "specify timeout value(unit: second), default to 3")
	flag.StringVar(&pacFile, "pac", "proxy.pac", "specify pac file, default to proxy.pac")
	flag.Parse()

	pacCode, err := ioutil.ReadFile(pacFile)
	if err != nil {
		panic(err)
	}

	pac.Parser.LoadPac(string(pacCode))

	f, _ := os.OpenFile("cache.yap", os.O_RDWR|os.O_CREATE, 0644)
	cache.RecoverFromReader(f)

	conf := &socks5.Config{}
	conf.Dial = func(socks5Addr string, timeout int) func(context.Context, string, string) (net.Conn, error) {
		dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
		if err != nil {
			panic(err)
		}
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, _ := net.SplitHostPort(addr)
			if net.ParseIP(host) != nil {
				return dialer.Dial(network, addr)
			}
			target, err := net.DialTimeout(network, addr, time.Second*time.Duration(timeout))
			if err != nil {
				return dialer.Dial(network, addr)
			} else {
				return target, err
			}
		}
	}(socks5Addr, timeout)

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	go func(listenAddr string) {
		if err := server.ListenAndServe("tcp", listenAddr); err != nil {
			panic(err)
		}
	}(listenAddr)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel
	fmt.Println("hehe")
	cache.DumpToWriter(f)
	f.Close()
}
