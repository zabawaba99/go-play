package main

import (
	"errors"
	"log"

	d "github.com/fsouza/go-dockerclient"
)

type DockerClient struct {
	api *d.Client
}

func NewClient(dockerHost string) (*DockerClient, error) {
	if dockerHost == "" {
		return nil, errors.New("empty docker host")
	}

	api, err := d.NewClient(dockerHost)
	if err != nil {
		return nil, err
	}

	return &DockerClient{api: api}, nil
}

func NewTLSClient(dockerHost, dockerCertPath string) (*DockerClient, error) {
	if dockerHost == "" {
		return nil, errors.New("empty docker host")
	}

	if dockerCertPath == "" {
		return nil, errors.New("empty certificate path")
	}

	cert := dockerCertPath + "/cert.pem"
	key := dockerCertPath + "/key.pem"
	ca := dockerCertPath + "/ca.pem"

	api, err := d.NewTLSClient(dockerHost, cert, key, ca)
	if err != nil {
		return nil, err
	}

	return &DockerClient{api: api}, nil
}

func (dc *DockerClient) Run(image, tag string) error {
	opts := d.ListImagesOptions{}
	images, err := dc.api.ListImages(opts)
	if err != nil {
		return err
	}

	if tag == "" {
		tag = "latest"
	}

	fullName := image + ":" + tag
	var img *d.APIImages
	for _, i := range images {
		if i.ID == image || i.RepoTags[0] == fullName {
			img = &i
			break
		}
	}

	if img == nil {
		// pull image
		return nil
	}

	// create container
	container, err := dc.createContainer(img.ID)
	if err != nil {
		return nil
	}

	if err := dc.api.StartContainer(img.ID, &config); err != nil {
		log.Printf("err: %s\n", err)
	}
	return nil
}

func (dc *DockerClient) createContainer(image string) (*d.Container, error) {
	opts := d.CreateContainerOptions{
		Config: &d.Config{Image: image},
	}
	dc.api.CreateContainer(opts)
}

func (dc *DockerClient) startContainer(container *d.Container) error {
	config := d.HostConfig{PublishAllPorts: true}
	return dc.api.StartContainer(container.ID, hostConfig)
}
