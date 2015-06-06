package fxdocker

import (
	"github.com/fsouza/go-dockerclient"
	"time"
	"os"
)


type ContainerInspect struct {
	ID 					string 					`json:"id"`  // container ID
	APIContainer 		docker.APIContainers 	`json:"api_container"`
	InspectContainer 	*docker.Container 		`json:"inspect"`
	TopCommand 			docker.TopResult 		`json:"top_command"`
	Running 			bool                    `json:"running"`
}

var (
	AvailableContainers = make(map[string]ContainerInspect) // Local containers map ID -> ContainerInspect
	ContainersPerImage = make(map[string][]string) // Containers per Image with ImageName -> array of Container IDs
	dockerClient *docker.Client
)


func getContainerInfo(con docker.APIContainers) ContainerInspect {
	dock_inspect, _ := dockerClient.InspectContainer(con.ID)
	dock_top, _ := dockerClient.TopContainer(con.ID, "aux")

	return ContainerInspect{
		ID:con.ID,
		APIContainer: con,
		InspectContainer:dock_inspect,
		TopCommand:dock_top,
	}
}

func (fxd *FxDaemon) StartContainerInspector() {
	var (
		err error
		dock_containers []docker.APIContainers
		im_name string
		cont_id_contains bool
	)
	dockerClient, err = docker.NewClient(fxd.DockerEndpoint)
	if err != nil {
		fxd.OnError(err)
		os.Exit(1)
	}

	for {
		dock_containers, err = dockerClient.ListContainers(docker.ListContainersOptions{All: true})
		if err != nil {
			fxd.OnError(err)
		} else {
			for _, con := range dock_containers  {
				AvailableContainers[con.ID] = getContainerInfo(con)
				im_name = fxd.getImageNameById(AvailableContainers[con.ID].InspectContainer.Image)
				if len(im_name) > 0 {
					cont_id_contains = false
					for _, cid := range ContainersPerImage[im_name] {
						if cid == con.ID {
							cont_id_contains = true
							break
						}
					}
					if !cont_id_contains {
						ContainersPerImage[im_name] = append(ContainersPerImage[im_name], con.ID)
					}
				}
				time.Sleep(time.Millisecond * 20)
			}
		}
		time.Sleep(time.Second * 1)
	}
}