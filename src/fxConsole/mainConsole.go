package fxConsole

import (
	"fmt"
	"fxdocker"
	"strconv"
	"strings"
	"os"
	"crypto/md5"
	"github.com/Sirupsen/logrus"
	"encoding/hex"
	"lib"
	"encoding/json"
)

var (
	FlaxtonConfigFile = fmt.Sprintf("%s/.flaxton", os.Getenv("HOME"))
)

func RunArguments(args []string) {
	// Loading Configurations if file exists
	console_config := ConsoleConfig{}
	if _, err := os.Stat(FlaxtonConfigFile); os.IsNotExist(err) {
		logrus.Warn("Config File Doesent exisists, it will be created after authorization")
	} else {
		console_config.LoadConfig()
	}

	switch args[1] {
		case "-d", "daemon":
			{
				daemon := fxdocker.NewDaemon(fxdocker.DockerEndpoint, false)
				if len(console_config.DaemonID) == 0 {
					fmt.Println("Generating ID for this server")
					console_config.DaemonID = lib.RandomString(25)
					fmt.Println(console_config.DaemonID)
					console_config.SaveConfig()
				}
				daemon.AuthKey = console_config.Authorization
				next_args := args[1:]
				for i, arg := range args[1:]  {
					switch arg {
						case "-host":
							{
								daemon.ListenHost = next_args[i+1]
							}
						case "-balancer-port":
							{
								daemon.BalancerPort, _ = strconv.Atoi(next_args[i+1])
							}
						case "-offline":
							{
								daemon.Offline = true
							}
						case "help", "-h", "-help":
							{
								fmt.Println("This command is for Starting Flaxton daemon load balancer and daemon API server")
								fmt.Println("Formant: flaxton daemon <OPTIONS>")
								fmt.Println("OPTIONS:")
								fmt.Println("-host  : Listenning address for daemon server, example: 127.0.0.1:8888")
								fmt.Println("-balancer-port  : Network port for load balancing trafic across docker containers and child servers")
								fmt.Println("-offline  : If this parameter exists, then daemon will be working without flaxton.io server")
								os.Exit(1)
							}
					}
				}

				daemon.ID = console_config.DaemonID
				daemon_call := fxdocker.DaemonRegisterCall{DaemonID:daemon.ID, BalancerPort:daemon.BalancerPort, IP:daemon.ListenHost}
				d_call, _ := json.Marshal(daemon_call)
				request_error := lib.PostJson(fmt.Sprintf("%s/daemon", fxdocker.FlaxtonContainerRepo), d_call, nil, fmt.Sprintf("%s|%s", console_config.Authorization, daemon.ID))
				if request_error != nil {
					fmt.Println("Unable to register daemon: ", request_error.Error())
					os.Exit(1)
				}

				fmt.Println("Starting Flaxton Daemon on Address: ", daemon.ListenHost)
				daemon.Run()
			}

		case "-t", "transfer":
			{
				if len(console_config.Authorization) == 0 {
					fmt.Printf("For this operation you need to login using your flaxton.io creditailes")
					fmt.Printf("flaxton login -u <username> -p <password>")
					fmt.Printf("Or If you don't have acount please register here https://flaxton.io/register")
					return
				}
				var (
					container_id = ""
					repo_name = ""
					destination = ""
					need_to_run = false
				)
				next_args := args[1:]
				for i, arg := range args[1:] {
					switch strings.ToLower(arg) {
						case "-c":  // Container ID parameter
							{
								container_id = next_args[i+1]
							}
						case "-repo": // Repo Name
							{
								repo_name = next_args[i+1]
							}
						case "-host": // Destination Host
							{
								destination = next_args[i+1]
							}
						case "-run": // Need to run or not
							{
								if strings.ToLower(next_args[i+1]) == "yes" || strings.ToLower(next_args[i+1]) == "y" {
									need_to_run = true
								}
							}
						case "help", "-h", "-help":
							{
								fmt.Println("This command is for transfering container to another Flaxton Daemon")
								fmt.Println("Formant: flaxton transfer <OPTIONS>")
								fmt.Println("OPTIONS:")
								fmt.Println("-c  : Option for specifiing container ID to transfer, could be also running container")
								fmt.Println("-repo  : Option for specifiing Image repository name for resporing it in destination after transfer")
								fmt.Println("-host  : Destionation host address, example: 192.168.1.5:8888")
								fmt.Println("-run  : Turn on container after transfering or not, availbale options [Yes/No], [y/n], default is No")
								os.Exit(1)
							}
					}
				}

				fxdocker.TransferContainer(container_id, repo_name, destination, need_to_run, console_config.Authorization)
			}
		case "-l", "login":
			{
				var (
					username = ""
					password = ""
				)
				next_args := args[1:]
				for i, arg := range args[1:] {
					switch strings.ToLower(arg) {
					case "-u":
						{
							username = next_args[i+1]
						}
					case "-p":
						{
							hasher := md5.New()
							hasher.Write([]byte(next_args[i+1]))
							password = hex.EncodeToString(hasher.Sum(nil))
						}
					case "help", "-h", "-help":
						{
							fmt.Println("This command is for sign in to flaxton.io service for transfering and using containers on large clusters")
							fmt.Println("Formant: flaxton login <OPTIONS>")
							fmt.Println("OPTIONS:")
							fmt.Println("-u  : Username for Flaxton")
							fmt.Println("-p  : Password for Flaxton")
							os.Exit(1)
						}
					}
				}

				console_config.Authorization = fxdocker.FlaxtonConsoleLogin(username, password)
				console_config.Username = username
				console_config.SaveConfig()
				fmt.Println("Login Successful !")
			}

	}
}