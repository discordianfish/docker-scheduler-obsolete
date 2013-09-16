package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/dotcloud/docker"
	"github.com/dotcloud/docker/client"
)

const (
	proto      = "tcp"
	version    = 1.5
	volumeRoot = "/srv"
)

var (
	ErrContainerNotFound = errors.New("Container not found")
)

type Container struct {
	ID     string `json:"ID"`
	Config struct {
		Hostname   string `json:"hostname"`
		Domainname string `json:"domainname"`
	} `json:"config"`

	Volumes         map[string]string `json:"volumes"`
	Args            []string          `json:"args"`
	Image           string            `json:"image"`
	NetworkSettings struct {
		PortMapping struct {
			Tcp map[string]string
		} `json:"PortMapping"`
	} `json:"NetworkSettings"`
}

type Dock string

func (d Dock) Kill(product string, sn serviceName) error {
	log.Printf("Killing %s/%s", product, sn)
	container, err := d.findContainer(product, sn)
	if err != nil {
		return err
	}

	client := client.New(proto, string(d), version)
	_, _, err = client.Call("POST", "/containers/"+container.ID+"/kill", nil)
	return err
}

func (d Dock) Schedule(job *Job) error {
	log.Printf("Scheduling %v on %s", job, d)
	client := client.New(proto, string(d), version)

	binds := []string{}
	volumes := map[string]struct{}{}
	for _, volume := range job.Volumes {
		src := fmt.Sprintf("%s/%s", volumeRoot, volume)  // same for now
		dest := fmt.Sprintf("%s/%s", volumeRoot, volume) // ^^^^^^^^^^^^
		log.Printf("- Mapping %s:%s", src, dest)
		volumes[dest] = struct{}{}
		binds = append(binds, fmt.Sprintf("%s:%s", src, dest))
	}

	ports := []string{}
	for _, p := range job.Services {
		ports = append(ports, strconv.Itoa(p))
	}
	config := &docker.Config{
		Image:      job.Image,
		Cmd:        job.Args,
		Volumes:    volumes,
		PortSpecs:  ports,
		Hostname:   job.Product,
		Domainname: string(job.ServiceName()),
	}

	hostConfig := &docker.HostConfig{
		Binds: binds,
	}

	body, st, err := client.Call("POST", "/containers/create", config)
	if st == 404 {
		return fmt.Errorf("Image not found: %s", err)
	}
	if err != nil {
		return err
	}

	runResult := &docker.APIRun{}
	err = json.Unmarshal(body, runResult)
	if err != nil {
		return err
	}
	log.Printf("Created new container: %v", *runResult)

	if _, _, err = client.Call("POST", "/containers/"+runResult.ID+"/start", hostConfig); err != nil {
		return err
	}
	return nil

}

func (d Dock) GetJobs() ([]*Job, error) {
	containers, err := d.getContainers()
	if err != nil {
		return nil, err
	}
	jobs := []*Job{}
	for _, container := range containers {
		job, err := d.getJob(container.ID)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (d Dock) getJob(id string) (*Job, error) {
	container, err := d.getContainer(id)
	if err != nil {
		return nil, err
	}
	return JobFromContainer(container)
}

func (d Dock) findContainer(product string, sn serviceName) (*Container, error) {
	log.Printf("Looking for %s/%s", product, sn)
	containers, err := d.getContainers()
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		log.Printf("is %v it?", container.Config)
		if container.Config.Hostname == product && container.Config.Domainname == string(sn) {
			return container, nil
		}
	}
	return nil, ErrContainerNotFound
}

func (d Dock) getContainers() ([]*Container, error) {
	client := client.New(proto, string(d), version)

	body, _, err := client.Call("GET", "/containers/json", nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get containers: %s", err)
	}

	cs := []docker.APIContainers{}
	if err := json.Unmarshal(body, &cs); err != nil {
		return nil, err
	}

	containers := []*Container{}
	for _, c := range cs {
		container, err := d.getContainer(c.ID)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func (d Dock) getContainer(id string) (*Container, error) {
	client := client.New(proto, string(d), version)

	body, _, err := client.Call("GET", "/containers/"+id+"/json", nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get container details for %s: %s", id, err)
	}

	c := Container{}
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
