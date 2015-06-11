package fxdocker

import (
	"fmt"
	"ipTablesController"
	"encoding/json"
	"time"
	"lib"
	"os"
)

var (
	DockerEndpoint = "unix:///var/run/docker.sock"
	FlaxtonContainerRepo = getFlaxtonRepo()
	DockerRegistry = getDockerRegistry()
)

func getFlaxtonRepo() string {
	ret_val := "http://container.flaxton.io"
	if len(os.Getenv("FLAXTON_REPO")) > 0 {
		ret_val = os.Getenv("FLAXTON_REPO")
	}
	return ret_val
}

func getDockerRegistry() string {
	ret_val := "flaxton.io:5000"
	if len(os.Getenv("FLAXTON_REGISTERY")) > 0 {
		ret_val = os.Getenv("FLAXTON_REGISTERY")
	}
	return ret_val
}

type ErrorHandler func(error)

type ChildServer struct {
	IP 		string 		`json:"ip"`  // IP for balancing
	Port 	int 		`json:"port"`  // port to balance
}

type BalancerImageInfo struct {
	Port 		int    		`json:"port"`
	ImageName	string    	`json:"image_name"`
	ImagePort	int         `json:"image_port"`
}

type FxDaemon struct  {
	ID 					string                        	`json:"id"`			// Random generated string for identification specific Daemon
	Name				string                        	`json:"name"`
	IP					string                        	`json:"ip"`
	AuthKey				string                        	`json:"auth_key"`	// Username for authentication
	BalancerPortImages  map[int][]BalancerImageInfo     `json:"-"`// Port -> BalancerImageInfo
	BalancerPortChild  	map[int][]ChildServer       	`json:"-"`// Port -> Child Server
	ParentHost 			string                          `json:"parent_host"`		// Parent Balancer IP port EX. 192.168.1.10:8888
	PendingTasks 		[]lib.Task                    	`json:"pending_tasks"` 		// Pending tasks to be done
	DockerEndpoint 		string                          `json:"docker_endpoint"` 	// Docker daemon socket endpoint
	Offline 			bool                            `json:"offilne"` 			// if this parameter is true then Daemon will be running without flaxton.io communication
	OnError 			ErrorHandler                    `json:"-"`
}

func NewDaemon(docker_endpoint string, offline bool) (fxd FxDaemon) {
	fxd.OnError = func(err error) {}
	fxd.DockerEndpoint = docker_endpoint
	fxd.Offline = offline
	fxd.BalancerPortImages = make(map[int][]BalancerImageInfo)
	fxd.BalancerPortChild = make(map[int][]ChildServer)
	return
}

func (fxd *FxDaemon) Register() error {
	d_call, e := json.Marshal(fxd)
	if e != nil {
		return e
	}
	return lib.PostJson(fmt.Sprintf("%s/daemon", FlaxtonContainerRepo), d_call, nil, fmt.Sprintf("%s|%s", fxd.AuthKey, fxd.ID))
}

func (fxd *FxDaemon) Run() {

	if !fxd.Offline {
		go fxd.ParentNotifier()
		go fxd.RunTasks()
	}

	go fxd.StartContainerInspector()
	go fxd.StartImageInspector()
	go fxd.PortToAddressMapping()

	ipTablesController.InitRouting()
}

type DaemonNotify struct {
	Tasks 			[]lib.Task 			`json:"tasks"`
	Restart 		bool        		`json:"restart"`
}


type StatData struct {
	Containers 		map[string]ContainerInspect    `json:"containers"` // AvailableContainers variable
	Images			map[string]ImageInspect        `json:"images"`
	// TODO: Here should be other state information for sending to server
}

// Sending Notification to base server, and getting child list + tasks to be done
func (fxd *FxDaemon) ParentNotifier() {
	var (
		send_buf []byte
		marshal_error error
		request_error error
		notify DaemonNotify
		state = StatData{}
		snd_map = make(map[string]StatData)
	)

	for {
		state.Containers = AvailableContainers
		state.Images = AvailableImages
		snd_map["state"] = state
		send_buf, marshal_error = json.Marshal(snd_map)
		if marshal_error != nil {
			fxd.OnError(marshal_error)
			return
		}

		request_error = lib.PostJson(fmt.Sprintf("%s/notify", FlaxtonContainerRepo), send_buf, &notify, fmt.Sprintf("%s|%s", fxd.AuthKey, fxd.ID))
		if request_error != nil {
			lib.LogError("Error from Parent Notifier request service", request_error)
		} else {
			// Reloading Task list ... Server should send task list once
			if len(notify.Tasks) > 0 {
				fxd.PendingTasks = append(fxd.PendingTasks, notify.Tasks...)
			}
		}

		time.Sleep(time.Second * 1)
	}
}