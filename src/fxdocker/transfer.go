package fxdocker

import (
	"net/http"
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"fmt"
	"os"
	"log"
	"encoding/json"
	"io/ioutil"
	"lib"
	"strings"
)

var (
	FlaxtonLoginUrl = fmt.Sprintf("%s/user/login", FlaxtonContainerRepo)
)

func FlaxtonConsoleLogin(username, password string) string {
	fmt.Println("Sending request to ", FlaxtonLoginUrl)
	json_strB := []byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, username, password))
	fmt.Println(string(json_strB))
	req, err := http.NewRequest("POST", FlaxtonLoginUrl, bytes.NewBuffer(json_strB))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
		case 200:
			{
				fmt.Println("Username and Password validated !")
			}
		case 401:
			{
				fmt.Println("Invalid Username or Password")
				fmt.Println("If you don't have account please register it here http://flaxton.io")
				os.Exit(1)
			}
		default:
			{
				fmt.Println("Server Response Error")
				os.Exit(1)
			}
	}

	body, read_error := ioutil.ReadAll(resp.Body)
	if read_error != nil {
		fmt.Println("Error Reading Response Body ! ")
		panic(read_error)
		os.Exit(1)
	}
	fmt.Println(string(body))
	var authorization_cb interface{}
	json_error := json.Unmarshal(body, &authorization_cb)
	if json_error != nil {
		fmt.Println("Error Reading Response Body !")
		panic(json_error)
		os.Exit(1)
	}

	auth_map := authorization_cb.(map[string]interface{})
	if _, ok := auth_map["authorization"]; !ok {
		fmt.Println("authorization key dosen't exisits in response !")
		os.Exit(1)
	}

	return auth_map["authorization"].(string)
}

func TransferImage(image, daemon, run_cmd, cpu_share, mem_set, run_count string, authorization string) {
	client, _ := docker.NewClient(DockerEndpoint)
	var (
		err error
		task_resp lib.TaskSendResponse
		image_names []string
		reg_image string
	)
	image_names = strings.Split(image, ":")
	reg_image = fmt.Sprintf("%s/%s", DockerRegistry, image_names[0])
	err = client.TagImage(image, docker.TagImageOptions{Repo:reg_image, Tag: image_names[1]})
	if err != nil {
		fmt.Println("Error Tagging Image: ", reg_image)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	err = client.PushImage(docker.PushImageOptions{
		Name:  reg_image,
		Registry: DockerRegistry,
		Tag: image_names[1],
		OutputStream: os.Stdout,
	}, docker.AuthConfiguration{Username:"test",Password:"test",ServerAddress:DockerRegistry})
	if err != nil {
		fmt.Println("Error Pushing Image to Registery: ", DockerRegistry)
		fmt.Println(err.Error())
		client.RemoveImage(fmt.Sprintf("%s:%s", reg_image, image_names[1]))
		os.Exit(1)
	}
	err = client.RemoveImage(fmt.Sprintf("%s:%s", reg_image, image_names[1]))
	if err != nil {
		fmt.Println("Error UnTagging : ", reg_image)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Image Pushed to registery")
	fmt.Println("Adding Task for Daemon: ", daemon)

	task_resp, err = AddTask(authorization, lib.TaskImageTransfer, daemon, map[string]string{
		"run_cmd": run_cmd,
		"image": image,
		"run_count": run_count,
		"cpu": cpu_share,
		"mem": mem_set,
	})

	if err != nil {
		fmt.Println("Error Adding Task to repository: ", FlaxtonContainerRepo)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Waiting task to be done: ", task_resp.TaskId, "\n")

	WaitTaskDone(task_resp.TaskId, authorization, func(){
		fmt.Print(".") // This function is colled on every request
	}, func(err error)bool{
		fmt.Println("Error on sending request:", err.Error())
		return false // if we will return true it will exit from Wait task
	}, func(t_res lib.TaskResult)bool {
		if t_res.Error {
			fmt.Println("Error Message from Task:", t_res.Message)
		}
		if t_res.Done {
			fmt.Println("Task Done:", task_resp.TaskId)
		}
		return true  // if this function is colled then we recieved taks marked as done or error
	})
}