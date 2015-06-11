package fxdocker

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"lib"
	"log"
	"time"
	"encoding/json"
	"strconv"
	"os"
	"strings"
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
				case lib.TaskImageTransfer:
					{
						current_result = lib.TaskResult{
							TaskID:t.TaskID,
							StartTime:time.Now().UTC().Unix(),
							Error: false,
							Done: false,
							Message: "",
						}
						cont_call := make(map[string]string)
						err := t.ConvertData(&cont_call)
						if err != nil {
							current_result.Error = true
							current_result.Message = err.Error()
						} else {
							err = fxd.TransferImage(cont_call)
							if err != nil {
								current_result.Error = true
								current_result.Message = err.Error()
							} else {
								current_result.Done = true
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


func AddTask(auth, task_type, daemon string, data interface{}) (task_resp lib.TaskSendResponse, err error) {
	sdn_map := make(map[string]interface{})
	sdn_map["task_type"] = task_type
	sdn_map["daemon"] = daemon
	sdn_map["data"] = data
	var send_buf []byte
	send_buf, err = json.Marshal(sdn_map)
	if err != nil {
		return
	}
	err = lib.PostJson(fmt.Sprintf("%s/task/add", FlaxtonContainerRepo), send_buf, &task_resp, auth)
	return
}

func WaitTaskDone(task_id, auth string, onTik func(), onError func(error)bool, onDone func(lib.TaskResult)bool) (task_res lib.TaskResult, err error) {
	send_buf := []byte(fmt.Sprintf(`{"task_id": "%s"}`, task_id))
	for {
		err = lib.PostJson(fmt.Sprintf("%s/task", FlaxtonContainerRepo), send_buf, &task_res, auth)
		if err != nil {
			if onError != nil {
				if onError(err) {
					return
				}
			}
		} else {
			if task_res.Error || task_res.Done {
				if onDone != nil {
					if onDone(task_res) {
						return
					}
				}
			}
		}
		if onTik != nil {
			onTik()
		}
		time.Sleep(time.Second * 2)
	}
}

func (fxd *FxDaemon) TransferImage(container_cmd map[string]string) (err error) {
	client, _ := docker.NewClient(fxd.DockerEndpoint)
	var (
		image_names []string
		reg_image string
	)

	image_names = strings.Split(container_cmd["image"], ":")
	reg_image = fmt.Sprintf("%s/%s", DockerRegistry, image_names[0])

	err = client.PullImage(docker.PullImageOptions{
		Repository: reg_image,
		Registry: DockerRegistry,
		Tag: image_names[1],
		OutputStream:os.Stdout,
	}, docker.AuthConfiguration{Username:"test",Password:"test",ServerAddress:DockerRegistry})

	if err != nil {
		fmt.Println("Error pulling image: ", reg_image)
		fmt.Println(err.Error())
		return
	}

	err = client.TagImage(fmt.Sprintf("%s:%s", reg_image, image_names[1]), docker.TagImageOptions{Repo:image_names[0], Tag:image_names[1]})

	if err != nil {
		fmt.Println("Error Tagging image: ", image_names[0])
		fmt.Println(err.Error())
		client.RemoveImage(reg_image)
		return
	}

	err = client.RemoveImage(reg_image)

	if err != nil {
		fmt.Println("Error Rmoving image: ", reg_image)
		fmt.Println(err.Error())
		return
	}

	if len(container_cmd["run_cmd"]) > 0 {  // if container start command set then we need to create container and run it
		var (
			cont *docker.Container
			count = 1
		)
		if len(container_cmd["run_count"]) > 0 {
			count, err = strconv.Atoi(container_cmd["run_count"])
			if err != nil {
				count = 1
			}
		}
		fmt.Println("Running Containers", count)
		for i := 0; i < count; i++  {
			cont, err = client.CreateContainer(docker.CreateContainerOptions{
				Config: &docker.Config{Cmd: []string{container_cmd["run_cmd"]}, Image: container_cmd["image"],AttachStdin: false, AttachStderr: false, AttachStdout: false},
				HostConfig: &docker.HostConfig{},
			})
			if err != nil {
				return err
			}
			client.StartContainer(cont.ID, &docker.HostConfig{})
			fmt.Println(cont.ID)
		}
	}
	return nil
}
