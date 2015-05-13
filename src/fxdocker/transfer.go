package fxdocker

import (
	"net/http"
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"mime/multipart"
	"fmt"
	"os"
	"log"
)


func TransferContainer(container_id, repo_name, dest_host string) {
	client, _ := docker.NewClient(DockerEndpoint)
	container, error_inspect := client.InspectContainer(container_id)
	if error_inspect != nil {
		fmt.Println("Error Inspecting Container: %s", error_inspect.Error())
		os.Exit(1)
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("docker_image", "docker_image.tar.gz")
	if err != nil {
		fmt.Println("Unable to create form file: %s", err.Error())
		os.Exit(1)
	}

	export_error := client.ExportImage(docker.ExportImageOptions{Name: container.Image, OutputStream: part})
	if export_error != nil {
		fmt.Println("Error Exporting Container: %s", export_error)
		os.Exit(1)
	}

	fmt.Println("Container Exported Successfully")
	writer.WriteField("run_command", container.Config.Cmd[0])
	writer.WriteField("image_name", repo_name)
	err = writer.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	request, err2 := http.NewRequest("POST", fmt.Sprintf("http://%s/container/transfer", dest_host), body)
	if err2 != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Println("Making Post Request")
	http_client := &http.Client{}
	resp, err3 := http_client.Do(request)
	if err3 != nil {
		log.Fatal(err3)
	} else {
		fmt.Println("Reuqest Done, Reading Response")
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		fmt.Println(body)
	}
}