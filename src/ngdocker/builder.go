package ngdocker

import (
	"log"
	"strings"
	"os"
	"fmt"
	"bufio"
	"github.com/fsouza/go-dockerclient"
)

type NgDockerConfig struct {
	DockerFile string // Path to DockerFile
	NginxConfig string // Path to Nginx configuration file
	DockerEndpoint string // endpoint for Docker Daemon unix:///var/run/docker.sock
	DockerImageName string // Image Name for NgDocker - Default "flaxton/ngdocker:main"
}

func CreateNgDockerConfig() (conf NgDockerConfig) {
	conf.NginxConfig = "nginx.conf" // By default
	conf.DockerFile = "."
	conf.DockerImageName = "flaxton/ngdocker:main"
	return conf
}

func (ngDockerConf *NgDockerConfig) ParseArguments(args []string) {
	if args[1] != "ngdocker" {
		log.Panic("2nd Argument should be ngdocker, for building Nginx Docker Container")
		return
	}
	var (
		arg string
	)
	for _, arg = range args[1:] {
		arg = strings.ToLower(arg)
		switch  {
			case strings.Contains(arg, "--dockerfile="):
				{
					ngDockerConf.DockerFile = strings.Replace(arg, "--dockerfile=", "", -1)
				}
			case strings.Contains(arg, "--nginxconf="):
				{
					ngDockerConf.NginxConfig = strings.Replace(arg, "--nginxconf=", "", -1)
				}
		}
	}
}

func (ngDockerConf *NgDockerConfig) BuildFromConfig() error {
	fmt.Println("Checking availibility of Docker !")
	path := strings.ToLower(os.Getenv("PATH"))
	if !strings.Contains(path, "docker") {
		fmt.Println("Docker is not Availble")
		fmt.Print("Do you want to install it ? [Y/N]")

		for {
			reader := bufio.NewReader(os.Stdin)
			value_input, _ := reader.ReadString('\n')
			value_input = strings.Replace(value_input, "\n", "", -1)
			if strings.ToLower(value_input) == "n" || strings.ToLower(value_input) == "no" {
				fmt.Println("\n Flaxton at this point working only with Docker !")
				fmt.Println("Exiting ....")
				os.Exit(1)
			}

			if strings.ToLower(value_input) != "y" || strings.ToLower(value_input) != "yes" {
				fmt.Println("\n Wrong Combination. Please Type 'Y' or 'N' ")
				continue
			}
			break
		}

		InstallDocker()
	}

	client, _ := docker.NewClient(ngDockerConf.DockerEndpoint)

	err := client.BuildImage(docker.BuildImageOptions{
		Name: ngDockerConf.DockerImageName,
		Dockerfile: "Dockerfile",
		SuppressOutput: true,
		Pull: true,
		OutputStream: os.Stdout,
		InputStream: os.Stdin,
		ContextDir: ngDockerConf.DockerFile,
	})

	if err != nil {
		return err
	}

	imgs, img_err := client.ListImages(docker.ListImagesOptions{All: false})
	if img_err != nil {
		return img_err
	}
	var img docker.APIImages
	for _, i := range imgs {
		if len(i.RepoTags) == 1 && i.RepoTags[0] == ngDockerConf.DockerImageName {
			img = i
			break
		}
	}

	fmt.Println("Image Created: %s", img.ID)
	container, errx := client.CreateContainer(docker.CreateContainerOptions{
		Name: "ngdocker",
		Config: docker.Config{
			Image: img.ID,
			Cmd: "/bin/bash echo 'aaaaaaaaa'", // TODO: Add here Nginx install command
			AttachStdout: true,
			AttachStdin: true,
			AttachStderr: true,
		},
		HostConfig: docker.HostConfig{

		},
	})

	if errx != nil {
		return errx
	}

	im, e := client.CommitContainer(docker.CommitContainerOptions{
		Container: container.ID,
		Repository: ngDockerConf.DockerImageName,
		Message: "From Flaxton: Installing Nginx",
		Author: "Flaxton",
	})

	if e != nil {
		return e
	}

	client.ListContainers()

	return nil
}