package ipTablesController

import (
	"time"
	"strconv"
	"os"
	"os/signal"
	"syscall"
)

type IpListFunction func() []string

type IpTableRouting struct {
	ServerIpListCallback IpListFunction
	RoutingPort int
	RoutingPortStr string
	RoutingProtocol string
}

func InitRouting(protocol string, port int, f IpListFunction) IpTableRouting {
	return IpTableRouting{
		ServerIpListCallback: f,
		RoutingPort: port,
		RoutingProtocol: protocol,
		RoutingPortStr: strconv.Itoa(port),
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

	// Checking Servers
	go func(){
		for {
			server_ips = ip.ServerIpListCallback()
			time.Sleep(time.Second * 2)
		}
	} ()

	// Handle Exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGQUIT)
	go func() {
		_ = <-sigc
		// Delete all rules for routing
		for _, ip_addr := range server_ips  {
			tablesCMD.ClearForwardIp(ip.RoutingPortStr, ip_addr, ip.RoutingProtocol)
		}
		DisableForwarding()
		os.Exit(1)
	}()

	var ip_addr string

	// Starting IP tables forward routing
	for {
		for _, ip_addr = range server_ips  {
			if prev_ip != "" {
				tablesCMD.ClearForwardIp(ip.RoutingPortStr, prev_ip, ip.RoutingProtocol)
			}
			time.Sleep(time.Millisecond * 300)
			tablesCMD.ForwardIp(ip.RoutingPortStr, ip_addr, ip.RoutingProtocol)
			prev_ip = ip_addr
			time.Sleep(time.Millisecond * 300)
		}
		time.Sleep(time.Millisecond * 500)
	}
}