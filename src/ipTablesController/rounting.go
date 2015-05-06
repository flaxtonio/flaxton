package ipTablesController

import (
	"fmt"
	"time"
	"strconv"
)

type IpListFunction func() []string

type IpTableRouting struct {
	ServerIpListCallback IpListFunction
	RoutingPort int
	RoutingProtocol string
}

func InitRouting(protocol string, port int, f IpListFunction) IpTableRouting {
	return IpTableRouting{
		ServerIpListCallback: f,
		RoutingPort: port,
		RoutingProtocol: protocol,
	}
}

func (ip *IpTableRouting) StartRouting() {
	var (
		tablesCMD IpTables
		server_ips []string
		prev_ip string // Storing previous ip for clearing forward role
	)
	prev_ip = ""
	tablesCMD, _ = GetIpTables()
	// Getting IP list for servers
	server_ips = ip.ServerIpListCallback()
	fmt.Println(server_ips)
	// Checking Servers
	go func(){
		for {
			server_ips = ip.ServerIpListCallback()
			time.Sleep(time.Second * 1)
		}
	} ()

	// Starting IP tables forward routing
	for {
		for _, ip_addr := range server_ips  {
			if prev_ip != "" {
				tablesCMD.ClearForwardIp(strconv.Itoa(ip.RoutingPort), prev_ip, ip.RoutingProtocol)
			}
			tablesCMD.ForwardIp(strconv.Itoa(ip.RoutingPort), prev_ip, ip.RoutingProtocol)
			prev_ip = ip_addr
			time.Sleep(time.Millisecond * 100)
		}
	}
}