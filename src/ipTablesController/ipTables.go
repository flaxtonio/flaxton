package ipTablesController

/*

FOR PORT FORWARDING

sysctl net.ipv4.ip_forward=1

// For Remote routing
iptables -t nat -A PREROUTING -p tcp --dport port -j DNAT --to-destination ip:port

// For Local routing
iptables -t nat -A OUTPUT -m addrtype --src-type LOCAL --dst-type LOCAL -p tcp --sport [source_port] --dport [dest_port] -j DNAT --to-destination [ip]

// For adding forwarding
iptables -t nat -A POSTROUTING -j MASQUERADE

*/

import (
	"net"
	"os"
	"os/exec"
)

const (
	IpTablesCMD = "iptables"
)

type IpTables struct {
}

func EnableForwarding() error {
	cmd1 := exec.Command("sysctl", "net.ipv4.ip_forward=1") // Making IPV4 able to forward
	e := cmd1.Run()
	if e != nil {
		return e
	}
	// Adding prerouting on forward
	cmd2 := exec.Command(IpTablesCMD, "-t", "nat", "-A", "POSTROUTING", "-j", "MASQUERADE")
	e2 := cmd2.Run()
	if e2 != nil {
		return e2
	}
	return nil
}

func DisableForwarding() error {
	cmd1 := exec.Command("sysctl", "net.ipv4.ip_forward=0") // Making IPV4 able to forward
	e := cmd1.Run()
	if e != nil {
		return e
	}
	// Adding prerouting on forward
	cmd2 := exec.Command(IpTablesCMD, "-t", "nat", "-D", "POSTROUTING", "-j", "MASQUERADE")
	e2 := cmd2.Run()
	if e2 != nil {
		return e2
	}
	return nil
}

func GetIpTables() (IpTables, error) {
	tb := IpTables{}
	EnableForwarding()
	return tb, nil
}

func forward_role(local_port, remote_adr, protocol, add_delete string) (cmd *exec.Cmd) {
	cmd = exec.Command(IpTablesCMD, "-t", "nat", add_delete, "PREROUTING", "-p", protocol,
		"--dport", local_port, "-j", "DNAT", "--to-destination", remote_adr)
//	cmd = exec.Command(IpTablesCMD, "-t", "nat", add_delete, "OUTPUT", "-p", protocol,
//		"--dport", local_port, "-j", "DNAT", "--to-destination", remote_adr)
	return
}

func deny_role(local_port, remote_adr, protocol, add_delete string) (cmd *exec.Cmd) {
	if remote_adr == "0.0.0.0" {
		cmd = exec.Command(IpTablesCMD, add_delete, "OUTPUT", "-p", protocol,
			"--destination-port", local_port, "-s", remote_adr, "-j", "DROP")
	} else if local_port != "0" {
		cmd = exec.Command(IpTablesCMD, add_delete, "OUTPUT", "-p", protocol,
			"--destination-port", local_port, "-j", "DROP")
	} else {
		cmd = exec.Command(IpTablesCMD, add_delete, "OUTPUT", "-p", protocol,
			"-s", remote_adr, "-j", "DROP")
	}
	return
}

func (ip *IpTables) ForwardIp(local_port, remote_adr, protocol string) error {
	cmd := forward_role(local_port, remote_adr, protocol, "-A")
	err := cmd.Run()
	return err
}

func (ip *IpTables) ClearForwardIp(local_port, remote_adr, protocol string) error {
	cmd := forward_role(local_port, remote_adr, protocol, "-D")
	err := cmd.Run()
	return err
}

func (ip *IpTables) DenyRole(local_port, remote_adr, protocol string) error {
	cmd := deny_role(local_port, remote_adr, protocol, "-A")
	err := cmd.Run()
	return err
}

func (ip *IpTables) ClearDenyRole(local_port, remote_adr, protocol string) error {
	cmd := deny_role(local_port, remote_adr, protocol, "-D")
	err := cmd.Run()
	return err
}

func (ip *IpTables) ReCalculateRoles() {
}

func PrivateIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
