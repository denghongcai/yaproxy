package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/proxy"

	"github.com/codegangsta/cli"
	"github.com/denghongcai/yaproxy/cache"
	"github.com/denghongcai/yaproxy/pac"
	"github.com/denghongcai/yaproxy/socks5"
)

var socks5Addr string
var listenAddr string
var timeout int
var pacFile string

func main() {
	app := cli.NewApp()
	app.Name = "yaproxy"
	app.Usage = "automatic proxy before your actual socks5 proxy"
	app.Version = "0.1.2"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "socks5, s",
			Value:       "127.0.0.1:1080",
			Usage:       "specify socks5 server behind yaproxy, default to 127.0.0.1:1080",
			Destination: &socks5Addr,
		},
		cli.StringFlag{
			Name:        "listen, l",
			Value:       "127.0.0.1:10800",
			Usage:       "specify listen address, default to 127.0.0.1:10800",
			Destination: &listenAddr,
		},
		cli.StringFlag{
			Name:        "pac, p",
			Value:       "proxy.pac",
			Usage:       "specify pac file, default to ./proxy.pac",
			Destination: &pacFile,
		},
		cli.IntFlag{
			Name:        "timeout, t",
			Value:       2,
			Usage:       "specify timeout value(unit: second), default to 2",
			Destination: &timeout,
		},
	}
	app.Action = func(c *cli.Context) {
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
				fmt.Println(addr)
				host, _, _ := net.SplitHostPort(addr)
				if net.ParseIP(host) == nil {
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

	app.Run(os.Args)

}
