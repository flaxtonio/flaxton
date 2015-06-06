package fxdocker

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"net/http"
	"lib"
	"log"
	"strings"
	"errors"
	"time"
	"encoding/json"
	"strconv"
)

func (fxd *FxDaemon) RunTasks() {
	var (
		send_buf []byte
		marshal_error error
		request_error error
		current_tasks []lib.Task
		current_result lib.TaskResult
	)

	for  {
		if len(fxd.PendingTasks) > 0 {
			current_tasks = make([]lib.Task, 0)
			current_tasks = append(current_tasks, fxd.PendingTasks...)
			fxd.PendingTasks = make([]lib.Task, 0)  // Clearing tasks
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
						} else {
							_, err = fxd.TransferContainer(cont_call)
							if err != nil {
								current_result.Error = true
								current_result.Message = err.Error()
							} else {
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
					}
				case lib.TaskSetDaemonName:
					{
						name_map := make(map[string]string)
						err := t.ConvertData(&name_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							fxd.Name = name_map["name"]
							fxd.Register()
							current_result.EndTime = time.Now().UTC().Unix()
						}
					}
				case lib.TaskAddChildServer:
					{
						child_map := make(map[string]ChildServer)
						err := t.ConvertData(&child_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							var port int
							for k, ch :=range child_map {
								port, _ = strconv.Atoi(k)
								if _, ok := fxd.BalancerPortChild[port]; !ok {
									fxd.BalancerPortChild[port] = make([]ChildServer, 0)
								}
								fxd.BalancerPortChild[port] = append(fxd.BalancerPortChild[port], ch)
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
					}
				case lib.TaskAddBalancingImage:
					{
						im_map := make(map[string]BalancerImageInfo)
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							var port int
							for k, im :=range im_map {
								port, _ = strconv.Atoi(k)
								if _, ok := fxd.BalancerPortImages[port]; !ok {
									fxd.BalancerPortImages[port] = make([]BalancerImageInfo, 0)
								}
								fxd.BalancerPortImages[port] = append(fxd.BalancerPortImages[port], im)
							}
							current_result.EndTime = time.Now().UTC().Unix()
						}
					}
				case lib.TaskStartBalancerPort:
					{
						im_map := make(map[string]int)
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							fxd.StartBalancerPort(im_map["port"])
							current_result.EndTime = time.Now().UTC().Unix()
						}
					}
				case lib.TaskStopBalancerPort:
					{
						im_map := make(map[string]int)
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							fxd.StopBalancerPort(im_map["port"])
							current_result.EndTime = time.Now().UTC().Unix()
						}
					}
				case lib.TaskStartContainer:
					{
						im_map := make(map[string]docker.HostConfig) // ContainerID -> Container Host Config
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							message := ""
							for cid, conf :=range im_map {
								err = StartContainer(cid, &conf)
								if err != nil {
									message = fmt.Sprintf("%s|%s&", cid, err.Error(), message)
								}
							}
							if message != "" {
								current_result.Error = true
								current_result.Message = message
							} else {
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
					}
				case lib.TaskStopContainer:
					{
						im_map := make(map[string]uint) // ContainerID -> Stop Timeout
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							message := ""
							for cid, t :=range im_map {
								err = StopContainer(cid, t)
								if err != nil {
									message = fmt.Sprintf("%s|%s&", cid, err.Error(), message)
								}
							}
							if message != "" {
								current_result.Error = true
								current_result.Message = message
							} else {
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
					}
				case lib.TaskPauseContainer:
					{
						im_map := make([]string, 0) // ContainerID Array to pause
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							message := ""
							for _, cid :=range im_map {
								err = PauseContainer(cid)
								if err != nil {
									message = fmt.Sprintf("%s|%s&", cid, err.Error(), message)
								}
							}
							if message != "" {
								current_result.Error = true
								current_result.Message = message
							} else {
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
					}
				case lib.TaskCreateContainer:
					{
						im_map := make(map[string]docker.CreateContainerOptions) // ContainerName -> Creation Options
						err := t.ConvertData(&im_map)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							message := ""
							for cname, opts :=range im_map {
								err = CreateContainer(opts)
								if err != nil {
									message = fmt.Sprintf("%s|%s&", cname, err.Error(), message)
								}
							}
							if message != "" {
								current_result.Error = true
								current_result.Message = message
							} else {
								current_result.EndTime = time.Now().UTC().Unix()
							}
						}
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


func (fxd *FxDaemon) AddTask(task_type, daemon string, data interface{}) error {
	sdn_map := make(map[string]interface{})
	sdn_map["task_type"] = task_type
	sdn_map["daemon"] = daemon
	sdn_map["data"] = data
	send_buf, marshal_error := json.Marshal(sdn_map)
	if marshal_error != nil {
		return marshal_error
	}
	return lib.PostJson(fmt.Sprintf("%s/task/add", FlaxtonContainerRepo), send_buf, nil, fxd.AuthKey)
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
