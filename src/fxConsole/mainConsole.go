package fxConsole

import (
	"fmt"
	"fxdocker"
	"strconv"
)

func RunArguments(args []string) {
	switch args[1] {
		case "daemon":
			{
				daemon := fxdocker.FxDaemon{}
				daemon.DockerEndpoint = fxdocker.DockerEndpoint
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
					}
				}
				fmt.Println("Starting Flaxton Daemon on Address: %s", daemon.ListenHost)
				daemon.Run()
			}
		case "transfer":
			{
				fxdocker.TransferContainer(args[2], args[3], args[4])
			}
	}
}