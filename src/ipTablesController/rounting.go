package ipTablesController

import (
	"time"
	"os"
	"os/signal"
	"syscall"
	"strconv"
)

var (
	AvailableRoutings = make(map[int]PortRouting)
	IpCommandChan = make(chan int)
)

type PortRouting struct {
	Port int
	Rule TableRule
	Running bool
}

func InitRouting() {

	// Handle Exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGQUIT)


	tablesCMD, _ := GetIpTables()

	go func() {
		_ = <-sigc
		// Delete all rules for routing
		tablesCMD.ClearDNATRole()
		DisableForwarding()
		os.Exit(1)
	}()

	for {
		<- IpCommandChan
		time.Sleep(time.Millisecond * 500)
		tablesCMD.RecalculateDNATRole()
		IpCommandChan <- 1
	}
}

func ReplaceRouting(port int, address string) {
	AvailableRoutings[port] = PortRouting{
		Port: port,
		Running: true,
		Rule: TableRule{
			Protocol: "tcp",
			LocalPort: strconv.Itoa(port),
			RemoteAddr: address,
		},
	}
}