package main

import (
    "fmt"
    "ipTablesController"
    "github.com/fsouza/go-dockerclient"
)

func main() {
    endpoint := "unix:///var/run/docker.sock"
    client, _ := docker.NewClient(endpoint)
//    fmt.Println(client.Info())
//    imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})
//    for _, img := range imgs {
//        fmt.Println("ID: ", img.ID)
//        fmt.Println("RepoTags: ", img.RepoTags)
//        fmt.Println("Created: ", img.Created)
//        fmt.Println("Size: ", img.Size)
//        fmt.Println("VirtualSize: ", img.VirtualSize)
//        fmt.Println("ParentId: ", img.ParentID)
//    }

    var callback = func() []string {
        ret_ips := make([]string, 0)
        conts, _ := client.ListContainers(docker.ListContainersOptions{All: false})
        for _, con := range conts {
//            fmt.Println("ID: ", con.ID)
//            client.RemoveContainer(docker.RemoveContainerOptions{ID: con.ID})
            inspect, _ := client.InspectContainer(con.ID)
            ret_ips = append(ret_ips, inspect.NetworkSettings.IPAddress)
		fmt.Println(inspect.NetworkSettings.IPAddress)
        }
        return ret_ips
    }
    fmt.Println(callback())
    ipRouting := ipTablesController.InitRouting("tcp", 80,callback)
    ipRouting.StartRouting()
}
