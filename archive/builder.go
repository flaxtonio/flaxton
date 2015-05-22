package fxdocker

import (
	"strings"
	"os"
	"fmt"
	"bufio"
	"github.com/fsouza/go-dockerclient"
)

type FlaxtonDockerConfig struct {
	DockerFile string // Path to DockerFile
	DockerEndpoint string // endpoint for Docker Daemon unix:///var/run/docker.sock
	DockerImageName string // Image Name for NgDocker - Default "flaxton/ngdocker:main"
}

func CreateFlaxtonDockerConfig() (conf FlaxtonDockerConfig) {
	conf.DockerFile = "."
	conf.DockerImageName = "main"
	return conf
}

func (ngDockerConf *FlaxtonDockerConfig) ParseArguments(args []string) {
	// If this function called then args[1] should be "build"
	// second argument probably should be Dockerfile context

	if len(args) >= 3 {
		ngDockerConf.DockerFile = args[2]
	}

	if len(args) >= 4 {
		ngDockerConf.DockerImageName = args[3]
	}
}

func (fxDockerConf *FlaxtonDockerConfig) BuildFromConfig() error {
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

	client, _ := docker.NewClient(fxDockerConf.DockerEndpoint)

	err := client.BuildImage(docker.BuildImageOptions{
		Name: fmt.Sprintf("flaxton/balancer:%s", fxDockerConf.DockerImageName),
		Dockerfile: "Dockerfile",
		SuppressOutput: true,
		Pull: true,
		OutputStream: os.Stdout,
		InputStream: os.Stdin,
		ContextDir: fxDockerConf.DockerFile,
	})

	if err != nil {
		return err
	}

	fmt.Println("Your Container Installed !")
	fmt.Println("Run command flaxton ")
	return nil
}