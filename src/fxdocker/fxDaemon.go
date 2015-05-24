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
)

const (
	DockerEndpoint = "unix:///var/run/docker.sock"
	FlaxtonContainerRepo = "http://container.flaxton.io"
)

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

type FxDaemon struct  {
	ID string				// Random generated string for identification specific Daemon
	ListenHost string 		// IP and port server to listen container requests EX. 0.0.0.0:8888
	Authentication bool 	// Authentication enabled or not
	AuthKey	string			// Username for authentication
	ParentHost string 	// Parent Balancer IP port EX. 192.168.1.10:8888
	BalancerPort int		// Port for Load Balancer
	child_servers []string // Only for Private use
	DockerEndpoint string // Docker daemon socket endpoint
	Offline bool // if this parameter is true then Daemon will be running without flaxton.io communication
	OnError ErrorHandler
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

// For Internal API calls
func (fxd *FxDaemon) AddChildServer(addr string) {
	contains := false
	for _, ch := range fxd.child_servers  {
		if ch == addr {
			contains = true
			break
		}
	}
	if !contains {
		fxd.child_servers = append(fxd.child_servers, addr)
	}
}

func (fxd *FxDaemon) DeleteChildServer(addr string) {
	var new_ips []string
	for _, ch := range fxd.child_servers  {
		if ch != addr {
			new_ips = append(new_ips, ch)
		}
	}
	fxd.child_servers = append(new_ips, []string{}...)
}

type ChildServersCall struct {
	Servers []string  `json:"servers"`
}

func (fxd *FxDaemon) ChildServerGetter() {
	var (
		send_buf  = []byte("{}")
		request_error error
		child_servers ChildServersCall
	)

	for {
		request_error = lib.PostJson(fmt.Sprintf("%s/childs", FlaxtonContainerRepo), send_buf, &child_servers)
		if request_error != nil {
			lib.LogError("Error from Child server getter", request_error)
		}
		fxd.child_servers = make([]string, 0)
		fxd.child_servers = append(fxd.child_servers, child_servers.Servers...)
		time.Sleep(time.Second * 2)
	}
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
		ret_ips = append(ret_ips, cs)
	}

	return ret_ips
}

func (fxd *FxDaemon) Run() {

	go fxd.containerInspector()

	ipRouting := ipTablesController.InitRouting("tcp", fxd.BalancerPort, fxd.ipTablesCallback)
	go ipRouting.StartRouting()

	if !fxd.Offline {
		go fxd.ParentNotifier()
		go fxd.ChildServerGetter()
	}

	// Start API server
	daemon_api := FxDaemonApi{Fxd: fxd}
	daemon_api.RunApiServer()
}

// TODO: ADD TASK STACK !!!

// Sending Notification to base server
func (fxd *FxDaemon) ParentNotifier() {
	var (
		send_buf []byte
		marshal_error error
		request_error error
	)

	for {
		send_buf, marshal_error = json.Marshal(containers)
		if marshal_error != nil {
			fxd.OnError(marshal_error)
			return
		}
		request_error = lib.PostJson(fmt.Sprintf("%s/put", FlaxtonContainerRepo), send_buf, nil)
		if request_error != nil {
			lib.LogError("Error from Parent Notifier request service", request_error)
		}
		time.Sleep(time.Second * 1)
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