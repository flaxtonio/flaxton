package fxdocker
import (
	"bytes"
	"os/exec"
	"strings"
	"strconv"
	"fmt"
	"os"
)


func InstallDocker() {
	distro, _, kernel := GetLinuxInfo()
	switch distro {
		case "ubuntu":
			{
				// Checking Kernel Version, Example: 3.19.0-16-generic
				version_splits := strings.Split(kernel, ".")
				g_version, _ := strconv.Atoi(version_splits[0])
				s_version, _ := strconv.Atoi(version_splits[1])
				if g_version < 3 || s_version < 13 {
					fmt.Println("Your Kernel Version is %s , Docker requires min. 3.13 version.", kernel)
					fmt.Println("Updating Kernel !")
					var k_apt string
					if strings.Contains(kernel, "generic") {
						k_apt = "linux-image-generic-generic-trusty"
					} else {
						k_apt = "linux-image-generic-lts-trusty"
					}
					update_cmd := exec.Command("apt-get", "install", k_apt)
					update_cmd.Stdout = os.Stdout
					update_cmd.Stdin = os.Stdin
					update_cmd.Run()
				}

				install_cmd := exec.Command("apt-get", "install", "wget", "&&", "wget", "-qO-", "https://get.docker.com/", "|", "sh")
				install_cmd.Stdout = os.Stdout
				install_cmd.Stdin = os.Stdin
				install_cmd.Run()
			}
		default:
			{
				fmt.Println("We not supporting your OS yet ! :(")
				os.Exit(1)
			}

	}
}

func GetLinuxInfo() (distribution, version, kernel_version string) {
	// Getting Linux distribution
	var (
		version_buf bytes.Buffer
	)
	version_cmd := exec.Command("lsb_release", "-a")
	version_cmd.Stdout = &version_buf
	version_cmd.Run()
	version_details := strings.Split(strings.ToLower(version_buf.String()), "\n")
	for _, d := range version_details {
		switch {
		case strings.Contains(d, "distributor id:"):
			{
				distribution = strings.Replace(strings.Replace(d, "distributor id:", "", -1), " ", "", -1)
			}
		case strings.Contains(d, "release:"):
			{
				version = strings.Replace(strings.Replace(d, "distributor id:", "", -1), " ", "", -1)
			}
		}
	}
	version_buf = bytes.Buffer{}
	version_cmd = exec.Command("uname", "-r")
	version_cmd.Stdout = &version_buf
	version_cmd.Run()
	kernel_version = version_buf.String()
	return
}