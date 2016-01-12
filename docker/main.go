package main

import (
	"encoding/json"
	"fmt"
	dc "github.com/fsouza/go-dockerclient"
	"io/ioutil"
)

type Config struct {
	Host     string
	CertPath string
}

type Data struct {
	DockerImages map[string]*dc.APIImages
}

type CreateContainer struct {
	Provider      Provider
	ContainerOpts ContainerOpts
}

type Provider struct {
	Host     string
	CertPath string
}

type ContainerOpts struct {
	Name     string
	Image    string
	Hostname string
	Command  []string
}

func (c *Config) NewClient() (*dc.Client, error) {
	return dc.NewClient(c.Host)
}

func fetchLocalImages(data *Data, client *dc.Client) error {
	images, err := client.ListImages(dc.ListImagesOptions{All: false})
	if err != nil {
		return fmt.Errorf("Unable to list Docker images: %s", err)
	}

	if data.DockerImages == nil {
		data.DockerImages = make(map[string]*dc.APIImages)
	}
	for i, image := range images {
		data.DockerImages[image.ID[:12]] = &images[i]
		data.DockerImages[image.ID] = &images[i]
		for _, repotag := range image.RepoTags {
			data.DockerImages[repotag] = &images[i]
		}
	}
	return nil
}

func ReadConfig(cc *CreateContainer) {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return
	}

	err = json.Unmarshal(file, &cc)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func ImageName(opts *dc.PullImageOptions, image string) {
	opts.Repository, opts.Tag = dc.ParseRepositoryTag(image)
}

func main() {
	var cc CreateContainer

	ReadConfig(&cc)
	fmt.Println(cc)

	var opts dc.PullImageOptions
	ImageName(&opts, cc.ContainerOpts.Image)

	var auth dc.AuthConfiguration
	var image string
	var data Data

	config := Config{
		Host:     cc.Provider.Host,
		CertPath: "",
	}

	cont_opts := dc.CreateContainerOptions{
		Name: cc.ContainerOpts.Name,
		Config: &dc.Config{
			Hostname: cc.ContainerOpts.Hostname,
			Cmd:      cc.ContainerOpts.Command,
		},
		HostConfig: &dc.HostConfig{
			Privileged: true,
		},
	}

	client, err := config.NewClient()
	if err != nil {
		fmt.Printf("Error initializing Docker client: %s\n", err)
		return
	}

	err = client.Ping()
	if err != nil {
		fmt.Printf("Error pinging Docker server: %s\n", err)
		return
	}
	fmt.Printf("Sucessfully connection to host: %s\n", config.Host)

	image = cc.ContainerOpts.Image
	err = fetchLocalImages(&data, client)
	if err != nil {
		fmt.Println("Error fetching image list")
	}
	if _, ok := data.DockerImages[image]; !ok {
		err = client.PullImage(opts, auth)
		if err != nil {
			fmt.Println("Error pulling image :", err)
			return
		}
	}

	cont_opts.Config.Image = image

	var retContainer *dc.Container
	if retContainer, err = client.CreateContainer(cont_opts); err != nil {
		fmt.Printf("Unable to create container: %s\n", err)
		return
	}
	if retContainer == nil {
		fmt.Println("Returned container is nil")
		return
	}
	if err = client.StartContainer(retContainer.ID, cont_opts.HostConfig); err != nil {
		fmt.Printf("Unable to start container: %s\n", err)
	}
	fmt.Println(retContainer.ID)
}
