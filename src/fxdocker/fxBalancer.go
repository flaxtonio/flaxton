package fxdocker

import (
	"ipTablesController"
	"lib"
	"time"
	"fmt"
)


var (
	BalancerPortStack = make(map[int]*lib.Stack)
	stop_port_chanels = make(map[int]chan bool)
	push_chanel = make(chan int)
)

func (fxd *FxDaemon) StartBalancerPort(port int) {
	if _, ok := BalancerPortStack[port]; ok {
		return // Balancer already running for this port
	}

	BalancerPortStack[port] = new(lib.Stack)
	stop_port_chanels[port] = make(chan bool)

	exit := false

	go func() {
		_ = <- stop_port_chanels[port]
		exit = true
	}()

	for {
		<- push_chanel
		if exit {
			delete(BalancerPortStack, port)
			delete(stop_port_chanels, port)
			delete(ipTablesController.AvailableRoutings, port)
			return
		}

		for BalancerPortStack[port].Len() > 0 {
			ipTablesController.ReplaceRouting(port, BalancerPortStack[port].Pop().(string))
			ipTablesController.IpCommandChan <- 1
			<- ipTablesController.IpCommandChan
		}
		push_chanel <- 1
	}
}

func (fxd *FxDaemon) StopBalancerPort(port int) {
	stop_port_chanels[port] <- true
}

func (fxd *FxDaemon) PortToAddressMapping() {
	var (
		ip_addr string
	)

	// Starting available ports for this moment
	for port, _ :=range fxd.BalancerPortImages {
		go fxd.StartBalancerPort(port)
	}

	// Starting available ports for this moment
	for port, _ :=range fxd.BalancerPortChild {
		go fxd.StartBalancerPort(port)
	}


	for {
		// Mapping Containers per Image
		for port, images :=range fxd.BalancerPortImages {
			for _, balancer_info := range images {
				for _, cont_id := range ContainersPerImage[balancer_info.ImageName] {
					ip_addr = AvailableContainers[cont_id].InspectContainer.NetworkSettings.IPAddress
					if len(ip_addr) == 0 || ip_addr == "" {
						continue
					}
					if _, ok := BalancerPortStack[port]; ok {
						BalancerPortStack[port].Push(fmt.Sprintf("%s:%d", ip_addr, balancer_info.Port))
					}
					time.Sleep(time.Millisecond * 10)
				}
			}
		}

		// Mapping Child Servers per port
		for port, childs :=range fxd.BalancerPortChild  {
			for _, child :=range childs  {
				ip_addr = child.IP
				if _, ok := BalancerPortStack[port]; ok {
					BalancerPortStack[port].Push(fmt.Sprintf("%s:%d", ip_addr, child.Port))
				}
				time.Sleep(time.Millisecond * 10)
			}
		}
		push_chanel <- 0
		<- push_chanel
	}
}