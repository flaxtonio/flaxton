package main

import (
    "ipTablesController"
    "github.com/fsouza/go-dockerclient"
    "os"
    "fmt"
)

func main() {
    endpoint := "unix:///var/run/docker.sock"
    client, _ := docker.NewClient(endpoint)
    imgs, img_err := client.ListImages(docker.ListImagesOptions{All: false})
    if img_err != nil {
        fmt.Println(img_err.Error())
    }
    fmt.Println(imgs[0].RepoTags)
    os.Exit(1)
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
//		fmt.Println(inspect.NetworkSettings.IPAddress)
        }
        return ret_ips
    }
    ipRouting := ipTablesController.InitRouting("tcp", 80,callback)
    ipRouting.StartRouting()
}
