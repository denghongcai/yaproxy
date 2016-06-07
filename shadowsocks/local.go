package shadowsocks

import (
	"github.com/spance/suft/protocol"
	"log"
	"math/rand"
	"net"
	"strconv"

	ss "github.com/denghongcai/shadowsocks-go/shadowsocks"
)

type ServerCipher struct {
	server string
	cipher *ss.Cipher
}

var servers struct {
	srvCipher       []*ServerCipher
	failCnt         []int // failed connection count
	srvSuftEndpoint []*suft.Endpoint
}

func ParseServerConfig(config *ss.Config) {
	hasPort := func(s string) bool {
		_, port, err := net.SplitHostPort(s)
		if err != nil {
			return false
		}
		return port != ""
	}

	suftParams := suft.Params{}

	if len(config.SuftParams) != 0 {
		suftParams = suft.Params{}
		suftParams.FastRetransmit = true
		suftParams.FlatTraffic = true
		suftParams.Bandwidth = int64(config.SuftParams["bandwidth"].(float64))
		log.Printf("enable suft protocol, bandwidth: %d\nMbps", suftParams.Bandwidth)
	}

	if len(config.ServerPassword) == 0 {
		// only one encryption table
		cipher, err := ss.NewCipher(config.Method, config.Password)
		if err != nil {
			log.Fatal("Failed generating ciphers:", err)
		}
		srvPort := strconv.Itoa(config.ServerPort)
		srvArr := config.GetServerArray()
		n := len(srvArr)
		servers.srvCipher = make([]*ServerCipher, n)
		servers.srvSuftEndpoint = make([]*suft.Endpoint, n)

		for i, s := range srvArr {
			if hasPort(s) {
				log.Println("ignore server_port option for server", s)
				servers.srvCipher[i] = &ServerCipher{s, cipher}
			} else {
				servers.srvCipher[i] = &ServerCipher{net.JoinHostPort(s, srvPort), cipher}
			}
			suftEndpoint, _ := suft.NewEndpoint(&suftParams)
			servers.srvSuftEndpoint[i] = suftEndpoint
		}
	} else {
		// multiple servers
		n := len(config.ServerPassword)
		servers.srvCipher = make([]*ServerCipher, n)
		servers.srvSuftEndpoint = make([]*suft.Endpoint, n)

		cipherCache := make(map[string]*ss.Cipher)
		i := 0
		for _, serverInfo := range config.ServerPassword {
			if len(serverInfo) < 2 || len(serverInfo) > 3 {
				log.Fatalf("server %v syntax error\n", serverInfo)
			}
			server := serverInfo[0]
			passwd := serverInfo[1]
			encmethod := ""
			if len(serverInfo) == 3 {
				encmethod = serverInfo[2]
			}
			if !hasPort(server) {
				log.Fatalf("no port for server %s\n", server)
			}
			cipher, ok := cipherCache[passwd]
			if !ok {
				var err error
				cipher, err = ss.NewCipher(encmethod, passwd)
				if err != nil {
					log.Fatal("Failed generating ciphers:", err)
				}
				cipherCache[passwd] = cipher
			}
			servers.srvCipher[i] = &ServerCipher{server, cipher}
			suftEndpoint, _ := suft.NewEndpoint(&suftParams)
			servers.srvSuftEndpoint[i] = suftEndpoint
			i++
		}
	}
	servers.failCnt = make([]int, len(servers.srvCipher))
	for _, se := range servers.srvCipher {
		log.Println("available remote server", se.server)
	}
	return
}

func connectToServer(serverId int, addr string) (remote *ss.Conn, err error) {
	se := servers.srvCipher[serverId]
	remote, err = ss.Dial(addr, se.server, se.cipher.Copy(), servers.srvSuftEndpoint[serverId])
	if err != nil {
		log.Println("error connecting to shadowsocks server:", err)
		const maxFailCnt = 30
		if servers.failCnt[serverId] < maxFailCnt {
			servers.failCnt[serverId]++
		}
		return nil, err
	}
	log.Printf("connected to %s via %s\n", addr, se.server)
	servers.failCnt[serverId] = 0
	return
}

func createServerConn(addr string) (remote *ss.Conn, err error) {
	const baseFailCnt = 20
	n := len(servers.srvCipher)
	skipped := make([]int, 0)
	for i := 0; i < n; i++ {
		// skip failed server, but try it with some probability
		if servers.failCnt[i] > 0 && rand.Intn(servers.failCnt[i]+baseFailCnt) != 0 {
			skipped = append(skipped, i)
			continue
		}
		remote, err = connectToServer(i, addr)
		if err == nil {
			return
		}
	}
	// last resort, try skipped servers, not likely to succeed
	for _, i := range skipped {
		remote, err = connectToServer(i, addr)
		if err == nil {
			return
		}
	}
	return nil, err
}

type LocalDialer struct{}

var Dialer LocalDialer

func (this LocalDialer) Dial(network, addr string) (net.Conn, error) {
	return createServerConn(addr)
}
