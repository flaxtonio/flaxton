package fxdocker


import (
	"github.com/fsouza/go-dockerclient"
	"time"
	"os"
)

type ImageInspect struct {
	ID 			string              `json:"id"`
	Name 		string              `json:"name"`  // Name is a combination of repository:tag from Docker
	ApiImage 	docker.APIImages 	`json:"api_image"`
	Inspect 	*docker.Image       `json:"inspect"`
}

var (
	AvailableImages = make(map[string]ImageInspect) // Available images Map with Name -> Image
)

func (fxd *FxDaemon) StartImageInspector() {
	var (
		err error
		dockerClient *docker.Client
		api_images []docker.APIImages
	)
	dockerClient, err = docker.NewClient(fxd.DockerEndpoint)
	if err != nil {
		fxd.OnError(err)
		os.Exit(1)
	}
	for  {
		api_images, err = dockerClient.ListImages(docker.ListImagesOptions{All:false})
		for _, img :=range api_images {
			if _, ok := AvailableImages[img.RepoTags[0]]; !ok {
				AvailableImages[img.RepoTags[0]] = ImageInspect{Name:img.RepoTags[0]}
			}
			im := AvailableImages[img.RepoTags[0]]
			im.ID = img.ID
			im.ApiImage = img
			im.Inspect, err = dockerClient.InspectImage(img.ID)
			if err != nil {
				fxd.OnError(err)
			}
			AvailableImages[img.RepoTags[0]] = im
			time.Sleep(time.Millisecond * 100)
		}
		time.Sleep(time.Second * 2)
	}
}

func (fxd *FxDaemon) getImageIdByName(name string) string {
	for n, im := range AvailableImages {
		if n == name {
			return im.ID
		}
	}
	return ""
}

func (fxd *FxDaemon) getImageNameById(id string) string {
	for n, im := range AvailableImages {
		if im.ID == id {
			return n
		}
	}
	return ""
}