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
				daemon.ID = console_config.DaemonID
				next_args := args[1:]
				for i, arg := range args[1:]  {
					switch arg {
						case "list":  // Get list of all daemons for current user
							{
								var list_resp []fxdocker.FxDaemon
								request_error := lib.PostJson(fmt.Sprintf("%s/daemon/list", fxdocker.FlaxtonContainerRepo), []byte("{}"), &list_resp, console_config.Authorization)
								if request_error != nil {
									fmt.Println("Response Error: ", request_error.Error())
									os.Exit(1)
								}
								fmt.Println("List of Daemons")
								for _, d := range list_resp {
									fmt.Println(d.ID, "     ", d.Name)
								}
								os.Exit(1)
							}
						case "-balancer":
							{
								local_port, _ := strconv.Atoi(next_args[i+1])
								image_port, _ := strconv.Atoi(next_args[i+3])
								daemon.BalancerPortImages[local_port] = make([]fxdocker.BalancerImageInfo, 0)
								daemon.BalancerPortImages[local_port] = append(daemon.BalancerPortImages[local_port], fxdocker.BalancerImageInfo{
									ImageName: next_args[i+2],
									ImagePort: image_port,
									Port: local_port,
								})
							}
						case "-child":
							{
								local_port, _ := strconv.Atoi(next_args[i+1])
								child_port, _ := strconv.Atoi(next_args[i+3])
								daemon.BalancerPortChild[local_port] = make([]fxdocker.ChildServer, 0)
								daemon.BalancerPortChild[local_port] = append(daemon.BalancerPortChild[local_port], fxdocker.ChildServer{
									IP: next_args[i+2],
									Port: child_port,
								})
							}
						case "-offline":
							{
								daemon.Offline = true
							}
						case "pause-container":
							{
								daemon_name := next_args[i+1]
								sdn_map := make([]string, 1)
								sdn_map[0] = next_args[i+2]
								task_err := daemon.AddTask(lib.TaskPauseContainer, daemon_name, sdn_map)
								if task_err != nil {
									fmt.Println(task_err)
								} else {
									fmt.Println("Task Sent !")
								}
								os.Exit(1)
							}
						case "set_name":
							{
								daemon_name := next_args[i+1]
								name := next_args[i+2]
								sdn_map := make(map[string]string)
								sdn_map["name"] = name
								task_err := daemon.AddTask(lib.TaskSetDaemonName, daemon_name, sdn_map)
								if task_err != nil {
									fmt.Println(task_err)
								} else {
									fmt.Println("Task Sent !")
								}
								os.Exit(1)
							}
						case "help", "-h", "-help":
							{
								fmt.Println("This command is for Starting Flaxton daemon load balancer and daemon API server")
								fmt.Println("Formant: flaxton daemon [COMMAND] <OPTIONS>")
								fmt.Println("COMMAND:")
								fmt.Println("list  : List of daemon servers for current logged in user")
								fmt.Println("OPTIONS:")
								fmt.Println("-balancer  : Options should be followed by this direction - local_port image_name image_port")
								fmt.Println("-child  : Add chaild server with this combination: [balancing port] [ip address]")
								fmt.Println("-offline  : If this parameter exists, then daemon will be working without flaxton.io server")
								os.Exit(1)
							}
					}
				}

				request_error := daemon.Register()
				if request_error != nil {
					fmt.Println("Unable to register daemon: ", request_error.Error())
					os.Exit(1)
				}

				fmt.Println("Starting Flaxton Daemon ! It will balance", daemon.ID, "port")
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
						case "-daemon": // Destination Host
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
								fmt.Println("-daemon  : Destionation Daemon Name or ID")
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