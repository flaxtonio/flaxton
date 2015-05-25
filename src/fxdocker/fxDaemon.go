package fxdocker

import (
	"net/http"
	"fmt"
	"ipTablesController"
	"github.com/fsouza/go-dockerclient"
	"encoding/json"
	"time"
	"strings"
	"log"
	"errors"
	"lib"
	"os"
)

var (
	DockerEndpoint = "unix:///var/run/docker.sock"
	FlaxtonContainerRepo = getFlaxtonRepo()
)

func getFlaxtonRepo() string {
	ret_val := "http://container.flaxton.io"
	if len(os.Getenv("FLAXTON_REPO")) > 0 {
		ret_val = os.Getenv("FLAXTON_REPO")
	}
	return ret_val
}

type ContainerInspect struct {
	ID string 							`json:"id"`  // container ID
	APIContainer docker.APIContainers 	`json:"api_container"`
	InspectContainer *docker.Container 	`json:"inspect"`
	TopCommand docker.TopResult 		`json:"top_command"`
}

var (
	dockerClient *docker.Client
	containers []ContainerInspect // Loac containers
)

type ErrorHandler func(error)

type ChildServer struct {
	IP string `json:"ip"`  // IP for balancing
	Port int `json:"port"`  // port to balance
}

type FxDaemon struct  {
	ID string				// Random generated string for identification specific Daemon
	ListenHost string 		// IP and port server to listen container requests EX. 0.0.0.0:8888
	Authentication bool 	// Authentication enabled or not
	AuthKey	string			// Username for authentication
	ParentHost string 	// Parent Balancer IP port EX. 192.168.1.10:8888
	BalancerPort int		// Port for Load Balancer
	child_servers []ChildServer // Only for Private use
	pendingTasks []lib.Task   // Pending tasks to be done
	DockerEndpoint string // Docker daemon socket endpoint
	Offline bool // if this parameter is true then Daemon will be running without flaxton.io communication
	OnError ErrorHandler
}

type DaemonListResponse struct {
	DaemonID string `json:"daemon_id"`
	Name string `json:"name"`
	IP string `json:"ip"`
	BalancerPort int `json:"balancer_port"`
	CreatedTime int64 `json:"created"` // Timestamp
	ChildServers []ChildServer `json:"child_servers"`
 }

type DaemonRegisterCall struct {
	DaemonID 		string 		`json:"daemon_id"`
	BalancerPort 	int 		`json:"balancer_port"`
	IP 				string 		`json:"ip"`
}

func NewDaemon(docker_endpoint string, offline bool) (fxd FxDaemon) {
	fxd.OnError = func(err error) {}
	fxd.DockerEndpoint = docker_endpoint
	fxd.Offline = offline
	return
}

func (fxd *FxDaemon) EnableAuthorization(key string) {
	fxd.Authentication = true
	fxd.AuthKey = key //TODO: Here should be some encryption
}

func (fxd *FxDaemon) containerInspector() {
	var (
		err error
		dock_containers []docker.APIContainers
		dock_inspect *docker.Container
		dock_top docker.TopResult
	)
	dockerClient, err = docker.NewClient(fxd.DockerEndpoint)
	if err != nil {
		fxd.OnError(err)
		return
	}
	for {
		dock_containers, err = dockerClient.ListContainers(docker.ListContainersOptions{All: false})
		if err != nil {
			fxd.OnError(err)
		} else {
			containers = make([]ContainerInspect, 0)
			for _, con := range dock_containers  {
				dock_inspect, _ = dockerClient.InspectContainer(con.ID)
				dock_top, _ = dockerClient.TopContainer(con.ID, "")
				containers = append(containers, ContainerInspect{
					ID:con.ID,
					APIContainer: con,
					InspectContainer:dock_inspect,
					TopCommand:dock_top,
				})
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func (fxd *FxDaemon) ipTablesCallback() []string {
	ret_ips := make([]string, 0)
	for _, con := range containers {
		ret_ips = append(ret_ips, con.InspectContainer.NetworkSettings.IPAddress)
	}

	for _, cs := range fxd.child_servers {
		ret_ips = append(ret_ips, fmt.Sprintf("%s:%d", cs.IP, cs.Port))
	}

	return ret_ips
}

func (fxd *FxDaemon) Run() {

	go fxd.containerInspector()

	if !fxd.Offline {
		go fxd.ParentNotifier()
		go fxd.RunTasks()
	}

	// Start API server
//	daemon_api := FxDaemonApi{Fxd: fxd}
//	daemon_api.RunApiServer()

	ipRouting := ipTablesController.InitRouting("tcp", fxd.BalancerPort, fxd.ipTablesCallback)
	ipRouting.StartRouting()
}

type DaemonNotify struct {
	ChildServers 	[]ChildServer  		`json:"child_servers"`
	Tasks 			[]lib.Task 			`json:"tasks"`
	Restart 		bool        		`json:"restart"`
}

// Sending Notification to base server, and getting child list + tasks to be done
func (fxd *FxDaemon) ParentNotifier() {
	var (
		send_buf []byte
		marshal_error error
		request_error error
		notify DaemonNotify
	)

	for {
		if len(containers) == 0 {
			send_buf = []byte("{}")  // empty array as json
		} else {
			send_buf, marshal_error = json.Marshal(containers)
			if marshal_error != nil {
				fxd.OnError(marshal_error)
				return
			}
		}

		request_error = lib.PostJson(fmt.Sprintf("%s/notify", FlaxtonContainerRepo), send_buf, &notify, fmt.Sprintf("%s|%s", fxd.AuthKey, fxd.ID))
		if request_error != nil {
			lib.LogError("Error from Parent Notifier request service", request_error)
		} else {
			// Adding child Servers
			fxd.child_servers = make([]ChildServer, 0)
			fxd.child_servers = append(fxd.child_servers, notify.ChildServers...)
			// Reloading Task list ... Server should send task list once
			if len(notify.Tasks) > 0 {
				fxd.pendingTasks = append(fxd.pendingTasks, notify.Tasks...)
			}
		}

		time.Sleep(time.Second * 1)
	}
}

func (fxd *FxDaemon) RunTasks() {
	var (
		send_buf []byte
		marshal_error error
		request_error error
		current_tasks []lib.Task
		current_result lib.TaskResult
	)

	for  {
		if len(fxd.pendingTasks) > 0 {
			current_tasks = make([]lib.Task, 0)
			current_tasks = append(current_tasks, fxd.pendingTasks...)
			fxd.pendingTasks = make([]lib.Task, 0)  // Clearing tasks
			fmt.Println("Task: Got new ", len(current_tasks), "tasks")
			// Starting execution for current tasks
			// TODO: maybe we will need concurrent execution in future
			for _, t := range current_tasks {
				fmt.Println("Task: ", t.TaskID, t.Type)
				switch t.Type {
					case lib.TaskContainerTransfer:
						{
							current_result = lib.TaskResult{
								TaskID:t.TaskID,
								StartTime:time.Now().UTC().Unix(),
								Error: false,
								Message: "",
							}
							cont_call := lib.TransferContainerCall{}
							err := t.ConvertData(&cont_call)
							if err != nil {
								current_result.Error = true
								current_result.Message = err.Error()
							}

							_, err = fxd.TransferContainer(cont_call)
							if err != nil {
								current_result.Error = true
								current_result.Message = err.Error()
							}
							current_result.EndTime = time.Now().UTC().Unix()
						}
					default:
						{
							current_result.TaskID = "-1"
						}
				}

				if current_result.TaskID == "-1" {
					fmt.Println("Unknown Task Command", t.Type)
				} else {
					send_buf, marshal_error = json.Marshal(current_result)
					if marshal_error != nil {
						log.Fatal(marshal_error)
					} else {
						request_error = lib.PostJson(fmt.Sprintf("%s/task", FlaxtonContainerRepo), send_buf, nil, fmt.Sprintf("%s|%s", fxd.AuthKey, fxd.ID))
						if request_error != nil {
							log.Fatal(request_error)
						} else {
							fmt.Println("Task: Done ", t.TaskID, current_result.EndTime)
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func (fxd *FxDaemon) TransferContainer(container_cmd lib.TransferContainerCall) (container_id string, err error) {
	fmt.Println("Getting Image from Flaxton Repo", FlaxtonContainerRepo)
	var resp *http.Response
	http_client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/images/%s", FlaxtonContainerRepo, container_cmd.ImageId), nil)
	req.Header.Set("Authorization", container_cmd.Authorization)
	resp, err = http_client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	fmt.Println("Loading Repo File to Docker Image Loader")
	client, _ := docker.NewClient(fxd.DockerEndpoint)
	err = client.LoadImage(docker.LoadImageOptions{InputStream: resp.Body})
	if err != nil {
		log.Panic(err.Error())
		return
	}

	image_name_split := strings.Split(container_cmd.ImageName, ":")

	fmt.Println("Making Sure That we are loaded image with Same ID", container_cmd.ImageId)
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
	img_found := false
	for _, img := range imgs {
		if img.ID == container_cmd.ImageId {
			img_found = true
		}
	}
	if !img_found {
		err = errors.New("Image Not Found After Loading it to Docker !")
		fmt.Println(err.Error())
		return "", err
	}

	fmt.Println("Image Found", container_cmd.ImageId)
	fmt.Println("Making Name Tagging", container_cmd.ImageName)
	client.TagImage(container_cmd.ImageId, docker.TagImageOptions{
		Repo: image_name_split[0],
		Tag: image_name_split[1],
		Force: true,
	})

	fmt.Println("Creating Container Based on Image", container_cmd.ImageId)
	var cont *docker.Container
	cont, err = client.CreateContainer(docker.CreateContainerOptions{
		//			Name: fmt.Sprintf("%s_%s", strings.Replace(strings.Replace(image_name, ":", "_", -1),"/","_", -1), "main"),
		Config: &docker.Config{Cmd: []string{container_cmd.Cmd}, Image: container_cmd.ImageId,AttachStdin: false, AttachStderr: false, AttachStdout: false},
		HostConfig: &docker.HostConfig{},
	})

	if err != nil {
		log.Panic(err.Error())
		return
	}

	container_id = cont.ID

	if container_cmd.NeedToRun {
		fmt.Println("Running Container", container_id)
		err = client.StartContainer(cont.ID, &docker.HostConfig{})
		if err != nil {
			log.Panic(err.Error())
			return
		}
	}

	return container_id, nil
}