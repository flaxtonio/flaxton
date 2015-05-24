package fxdocker

import (
	"net/http"
	"github.com/fsouza/go-dockerclient"
	"bytes"
	"mime/multipart"
	"fmt"
	"os"
	"log"
	"encoding/json"
	"io/ioutil"
	"github.com/cheggaaa/pb"
	"io"
	"lib"
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

func TransferContainer(container_id, repo_name, dest_host string, transfer_and_run bool, authorization string) {
	client, _ := docker.NewClient(DockerEndpoint)
	container, error_inspect := client.InspectContainer(container_id)
	if error_inspect != nil {
		fmt.Println("Error Inspecting Container: ", error_inspect.Error())
		os.Exit(1)
	}

	transfer_img := lib.TransferContainerCall{
		Name: container.Name,
		Cmd: container.Config.Cmd[0],
		ImageName: repo_name,
		ImageId: container.Image,
		NeedToRun: transfer_and_run,
		Authorization: authorization,
	}
	transfer_json_buf, convert_error := json.Marshal(transfer_img)
	json_length := bytes.Index(transfer_json_buf, []byte{0}) // For converting to string

	if convert_error != nil {
		fmt.Println("Error converting TransferContainerCall structure to json: ", convert_error.Error())
		os.Exit(1)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("docker_image", "docker_image.tar")
	if err != nil {
		fmt.Println("Unable to create form file: %s", err.Error())
		os.Exit(1)
	}

	fmt.Println("Exporting Image", transfer_img.ImageId)
	export_error := client.ExportImage(docker.ExportImageOptions{Name: container.Image, OutputStream: part})
	if export_error != nil {
		fmt.Println("Error Exporting Container: %s", export_error)
		os.Exit(1)
	}

	fmt.Println("Container Exported Successfully")
	writer.WriteField("image_info", string(transfer_json_buf[:json_length]))
	err = writer.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Adding progress bar for Uploading
	bar := pb.New(body.Len()).SetUnits(pb.U_BYTES)
	bar.Start()

	request, err2 := http.NewRequest("POST", fmt.Sprintf("%s/images/add", FlaxtonContainerRepo), io.TeeReader(body, bar))
	if err2 != nil {
		log.Fatal(err2)
		os.Exit(1)
	}

	request.Header.Set("Authorization", authorization)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Println("Making Post Request")

	http_client := &http.Client{}
	resp, err3 := http_client.Do(request)
	if err3 != nil {
		log.Fatal(err3)
		os.Exit(1)
	}

	fmt.Println("Reuqest Done, Reading Response")
	resp_body := &bytes.Buffer{}
	_, resp_err := resp_body.ReadFrom(resp.Body)
	if resp_err != nil {
		log.Fatal(resp_err)
		os.Exit(1)
	}

	resp.Body.Close()
	fmt.Println(body)
	fmt.Println("Sending Image Info to ", dest_host, " Destination host")

	// Making request with Image info to Destination host
	body2 := &bytes.Buffer{}
	body_bytes, marshal_error := json.Marshal(transfer_img)
	if marshal_error != nil {
		log.Fatal(marshal_error)
		os.Exit(1)
	}

	body2.Write(body_bytes)
	request, err2 = http.NewRequest("POST", fmt.Sprintf("http://%s/transfer/container", dest_host), body2)
	if err2 != nil {
		log.Fatal(err2)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err3 = http_client.Do(request)
	if err3 != nil {
		log.Fatal(err3)
		os.Exit(1)
	}

	fmt.Println("Container with Image sent successfully !")
	fmt.Println("Exiting !")
	os.Exit(1)
}