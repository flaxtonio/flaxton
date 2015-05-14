package fxdocker

import (
	"net/http"
	"fmt"
	"ipTablesController"
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"encoding/json"
	"time"
	"io/ioutil"
	"io"
	"strings"
)

const DockerEndpoint = "unix:///var/run/docker.sock"

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

	child_info_handler := func(w http.ResponseWriter, r *http.Request){
		fxd.AddChildServer(r.RemoteAddr)
	}

	// Start HTTP server for getting child IPs
	mux := http.NewServeMux()
	mux.HandleFunc("/", child_info_handler)
	fxd.StartAPIServer(mux)
}

// API Call structures
type ChildServersRequest struct {
	Command string `json:"command"`
	Servers []string `json:"servers"`
}

type DockerTransfer struct {
	RunCommand string `json:"run_command"`	// Runner command after transferring Docker container
	ImageName string  `json:"image_name"`	// Name for image, to define it after transfer
}

func (fxd *FxDaemon) StartAPIServer(mux *http.ServeMux) {
	add_new_child := func(w http.ResponseWriter, r *http.Request){
		var chs ChildServersRequest
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		if err := json.Unmarshal(body, &chs); err != nil {
			io.WriteString(w, err.Error())
			return
		}

		switch chs.Command {
			case "add":
				{
					for _, sip := range chs.Servers {
						fxd.AddChildServer(sip)
					}
				}
			case "delete":
				{
					for _, sip := range chs.Servers {
						fxd.DeleteChildServer(sip)
					}
				}
		}
	}

	transfer_container := func(w http.ResponseWriter, r *http.Request){
		fmt.Println("Parsing Request")
		r.ParseMultipartForm(32 << 20)
		command := r.FormValue("run_command")
		image_name := r.FormValue("image_name")

		fmt.Println("Reading Post File")
		file, _, err := r.FormFile("docker_image")
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer file.Close()

		client, _ := docker.NewClient(fxd.DockerEndpoint)
		fmt.Println("Loading Image")
		io.WriteString(w, "Loading Image")
		load_error := client.LoadImage(docker.LoadImageOptions{InputStream: file})
		if load_error != nil {
			io.WriteString(w, fmt.Sprintf("Error Loading Image: %s", load_error.Error()))
			return
		}

		imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
		last_created := imgs[0]
		for _, img := range imgs {
			fmt.Println(img.RepoTags[0], image_name, img.Created)
			if last_created.Created < img.Created {
				last_created = img
			}
		}
		fmt.Println(last_created.RepoTags[0], last_created.ID)
		cont, cont_error := client.CreateContainer(docker.CreateContainerOptions{
			Name: fmt.Sprintf("%s_%s", strings.Replace(strings.Replace(image_name, ":", "_", -1),"/","_", -1), "main"),
			Config: &docker.Config{Cmd: []string{command}, Image: last_created.RepoTags[0],AttachStdin: false, AttachStderr: false, AttachStdout: false},
			HostConfig: &docker.HostConfig{},
		})
		if cont_error != nil {
			io.WriteString(w, fmt.Sprintf("Error Creating Container: %s", cont_error.Error()))
			return
		}
		start_error := client.StartContainer(cont.ID, &docker.HostConfig{})
		if start_error != nil {
			io.WriteString(w, fmt.Sprintf("Error Starting Container: %s", start_error.Error()))
			return
		}
		io.WriteString(w, "Container Started")
		return
	}

	mux.HandleFunc("/childs", add_new_child)
	mux.HandleFunc("/container/transfer", transfer_container)
//	mux.HandleFunc("/container/stop", stop_container)
//	mux.HandleFunc("/container/start", start_container)

	http.ListenAndServe(fxd.ListenHost, mux)
}