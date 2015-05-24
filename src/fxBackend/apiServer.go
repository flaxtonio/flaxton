package fxBackend

import (
	"net/http"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"sync"
	"fmt"
	"os"
	"path"
	"encoding/json"
	"lib"
	"io"
)

var (
	ImagesDirectory = "./images"  // By default lets store images from transfer upload on the same directory
)

type FlaxtonBackend struct {
	ListenHost string
}

var (
	lock = sync.RWMutex{}
	//TODO: TACKS SHOULD BE IN MONGODB, for identification and task storage by Daemon ID
	tasks = TaskStack{}
)

func (back *FlaxtonBackend) RunApiServer() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/images/add", back.ContainerTransfer),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(back.ListenHost, api.MakeHandler()))
}

func (back *FlaxtonBackend) ContainerTransfer(w rest.ResponseWriter, r *rest.Request) {
	r.ParseMultipartForm(100000) // TODO: maybe will need concurrent file upload
	transfer_info_json := r.FormValue("image_info")
	var transfer_info lib.TransferContainerCall
	convert_error := json.Unmarshal([]byte(transfer_info_json), &transfer_info)
	if convert_error != nil {
		fmt.Fprintln(w, convert_error)
		log.Fatal(convert_error)
		return
	}

	file, _, err := r.FormFile("docker_image")
	if err != nil {
		fmt.Fprintln(w, err)
		log.Fatal(err)
		return
	}

	out, err := os.Create(path.Join(ImagesDirectory, transfer_info.ImageId))
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		log.Fatal("Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()
	defer file.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
		log.Fatal(err)
	}

	fmt.Fprintf(w, "File uploaded successfully")
	fmt.Println("File uploaded successfully")

	fmt.Fprintf(w, "Adding Container Information to Server Task List")
	fmt.Println("Adding Container Information to Server Task List")
	tasks.Add(lib.Task{
		ID: lib.RandomString(15),
		Data: transfer_info,
		Type: lib.TaskContainerTransfer,
		Cron: false,
	})
}