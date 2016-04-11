package main

import (
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"log"

	"golang.org/x/net/context"
	"golang.org/x/net/proxy"

	"github.com/codegangsta/cli"
	"github.com/denghongcai/yaproxy/cache"
	"github.com/denghongcai/yaproxy/gfwlist"
	"github.com/denghongcai/yaproxy/pac"
	"github.com/denghongcai/yaproxy/shadowsocks"
	"github.com/denghongcai/yaproxy/socks5"
	"github.com/denghongcai/yaproxy/util"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

var socks5Addr string
var listenAddr string
var timeout int
var logFile string
var pacFile string
var gfwListFile string
var ssConfigFile string

func main() {
	app := cli.NewApp()
	app.Name = "yaproxy"
	app.Usage = "automatic proxy before your actual socks5 proxy"
	app.Version = "0.3.3"
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
			Name:        "log",
			Value:       "yaproxy.log",
			Usage:       "specify log file, default to ./yaproxy.log",
			Destination: &logFile,
		},
		cli.StringFlag{
			Name:        "pac, p",
			Value:       "proxy.pac",
			Usage:       "specify pac file, default to ./proxy.pac",
			Destination: &pacFile,
		},
		cli.StringFlag{
			Name:        "gfwlist, gfw",
			Value:       "gfwlist.txt",
			Usage:       "specify gfwlist file, default to ./gfwlist.txt",
			Destination: &gfwListFile,
		},
		cli.StringFlag{
			Name:        "shadowsocks-config, ssc",
			Value:       "ss-config.json",
			Usage:       "specify shadowsocks config, default to ./ss-config.json",
			Destination: &ssConfigFile,
		},
		cli.IntFlag{
			Name:        "timeout, t",
			Value:       2,
			Usage:       "specify timeout value(unit: second), default to 2",
			Destination: &timeout,
		},
	}
	app.Action = func(c *cli.Context) {

		lf, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer lf.Close()
		log.SetOutput(lf)

		exists, err := ss.IsFileExists(pacFile)
		if exists && err == nil {
			pacCode, err := ioutil.ReadFile(pacFile)
			if err != nil {
				panic(err)
			}
			pac.Parser.LoadPac(string(pacCode))
		}

		exists, err = ss.IsFileExists(gfwListFile)
		if exists && err == nil {
			gfwListCode, err := ioutil.ReadFile(gfwListFile)
			if err != nil {
				panic(err)
			}
			gfwlist.Parser.LoadGFWList(string(gfwListCode))
		}

		f, _ := os.OpenFile("cache.yap", os.O_RDWR|os.O_CREATE, 0644)
		defer f.Close()
		cache.RecoverFromReader(f)

		conf := &socks5.Config{}
		conf.Dial = func(socks5Addr string, timeout int) func(context.Context, string, string) (net.Conn, error) {
			// socks5 forwarder
			dialer, err := proxy.SOCKS5("tcp", socks5Addr, nil, proxy.Direct)
			if err != nil {
				panic(err)
			}

			// shadowsocks forwarder
			exists, err := ss.IsFileExists(ssConfigFile)
			if exists && err == nil {
				config, err := ss.ParseConfig(ssConfigFile)
				if err == nil {
					shadowsocks.ParseServerConfig(config)
					dialer = shadowsocks.Dialer
				}
			} else {
				log.Println("shadowsocks config file not exists")
			}

			return func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, portString, _ := net.SplitHostPort(addr)
				port, _ := strconv.Atoi(portString)
				url := util.BuildURL(host, port)
				if net.ParseIP(host) == nil {
					target, err := dialer.Dial(network, addr)
					if err == nil {
						cache.AddURL(url, true)
					}
					return target, err
				}
				target, err := net.DialTimeout(network, addr, time.Second*time.Duration(timeout))
				if err != nil {
					target, err := dialer.Dial(network, addr)
					if err == nil {
						cache.AddURL(url, true)
					}
					return target, err
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
		log.Println("dump cached rules to file...")
		cache.DumpToWriter(f)
		log.Println("gracefully exit")
	}

	app.Run(os.Args)
}
