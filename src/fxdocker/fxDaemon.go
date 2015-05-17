package fxdocker

import (
	"net/http"
	"fmt"
	"ipTablesController"
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"encoding/json"
	"time"
	"strings"
	"log"
	"errors"
)

const (
	DockerEndpoint = "unix:///var/run/docker.sock"
	FlaxtonContainerRepo = "http://127.0.0.1:8080"
)

type FxDaemon struct  {
	ListenHost string 		// IP and port server to listen container requests EX. 0.0.0.0:8888
	Authentication bool 	// Authentication enabled or not
	AuthKey	string			// Username for authentication
	ParentHosts []string 	// Parent Balancer IP port EX. 192.168.1.10:8888
	BalancerPort int		// Port for Load Balancer
	child_servers []string // Only for Private use
	containers []docker.APIContainers // Loac containers
	DockerEndpoint string // Docker daemon socket endpoint
}

func (fxd *FxDaemon) EnableAuthorization(key string) {
	fxd.Authentication = true
	fxd.AuthKey = key //TODO: Here should be some encryption
}

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

func (fxd *FxDaemon) Run() {
	client, _ := docker.NewClient(fxd.DockerEndpoint)

	// Starting Balancer for Containers and child servers
	callback := func() []string {
		ret_ips := make([]string, 0)
		fxd.containers, _ = client.ListContainers(docker.ListContainersOptions{All: false})
		for _, con := range fxd.containers {
			inspect, _ := client.InspectContainer(con.ID)
			ret_ips = append(ret_ips, inspect.NetworkSettings.IPAddress)
		}

		for _, cs := range fxd.child_servers {
			ret_ips = append(ret_ips, cs)
		}

		return ret_ips
	}
	ipRouting := ipTablesController.InitRouting("tcp", fxd.BalancerPort,callback)
	go ipRouting.StartRouting()


	// Send State to Parents every 1 Second
	if len(fxd.ParentHosts) > 0 {
		go func() {
			var (
				cont_str []byte
				cont_err error
				req *http.Request
				req_err error
				client *http.Client
				resp *http.Response
				resp_error error
			)

			for {
				cont_err = nil
				req_err = nil
				resp_error = nil
				cont_str, cont_err = json.Marshal(fxd.containers)
				if cont_err == nil {
					for _, p := range fxd.ParentHosts  {
						req, req_err = http.NewRequest("POST", fmt.Sprintf("http://%s", p), bytes.NewBuffer(cont_str))
						if req_err != nil {
							fmt.Println(resp_error.Error())
							continue
						}
						req.Header.Set("Content-Type", "application/json")

						client = &http.Client{}
						resp, resp_error = client.Do(req)
						if resp_error != nil {
							panic(resp_error)
						}
						defer resp.Body.Close()
						// TODO: Maybe we will need to read parent response body !
					}
				}
				time.Sleep(time.Second * 1)
			}
		}()
	}

	// Start API server
	daemon_api := FxDaemonApi{Fxd: fxd}
	daemon_api.RunApiServer()
}

func (fxd *FxDaemon) TransferContainer(container_cmd TransferContainerCall) (container_id string, err error) {
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