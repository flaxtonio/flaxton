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
	"os/exec"
	"bytes"
	"fmt"
	"os"
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

type TableRule struct  {
	LocalPort string
	RemoteAddr string
	Protocol string
}


var restore_rule = "-A PREROUTING -p %s -m tcp --dport %s -j DNAT --to-destination %s"

// Table rule by balancing ports, every time they will change their accessable rule
func (ip *IpTables) RecalculateDNATRole() {
	if len(AvailableRoutings) == 0 {
		return
	}
	buffer := bytes.NewBufferString("*nat\n")
	cmd := exec.Command("iptables-restore", "--table=nat")
	for _, r := range AvailableRoutings  {
		buffer.WriteString(fmt.Sprintf(restore_rule, r.Rule.Protocol, r.Rule.LocalPort, r.Rule.RemoteAddr))
		buffer.WriteString("\n")
	}
	buffer.WriteString("COMMIT\n")
	cmd.Stdin = buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

var clear_rules_string = `
*nat
COMMIT
`

func (ip *IpTables) ClearDNATRole() {
	cmd := exec.Command("iptables-restore", "--table=nat")
	cmd.Stdin = bytes.NewBufferString(clear_rules_string)
	cmd.Run()
}
