package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type serviceName string

type Job struct {
	Image    string   `json:"image"`
	Args     []string `json:"args"`
	Volumes  []string `json:"volumes"`
	Services []int    `json:"services"`

	Product string `json:"product"`
	Env     string `json:"env"`
	Job     string `json:"job"`

	Docks []string `json:"docks"`
}

func JobFromFile(filename string) (*Job, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	job := &Job{}
	err = json.Unmarshal(file, &job)
	return job, err
}

func JobFromContainer(c *Container) (*Job, error) {
	volumes := []string{}
	for v, _ := range c.Volumes {
		volumes = append(volumes, v)
	}

	services := []int{}
	for private, _ := range c.NetworkSettings.PortMapping.Tcp {
		port, err := strconv.Atoi(private)
		if err != nil {
			return nil, err
		}
		services = append(services, port)
	}
	log.Printf("host: %s, domain: %s", c.Config.Hostname, c.Config.Domainname)
	serviceName := strings.Split(c.Config.Domainname, ".")
	if len(serviceName) != 2 {
		return nil, fmt.Errorf("Invalid container, couldn't derive service name from domain %s", c.Config.Domainname)
	}
	job := serviceName[0]
	env := serviceName[1]
	return &Job{
		Image:    c.Image,
		Args:     c.Args,
		Volumes:  volumes,
		Services: services,
		Product:  c.Config.Hostname,
		Env:      env,
		Job:      job,
	}, nil
}

func (j *Job) SameAs(otherJob *Job) bool {
	return j.Product == otherJob.Product &&
		j.Env == otherJob.Env &&
		j.Job == otherJob.Job
}

func (j *Job) ServiceName() serviceName {
	return serviceName(fmt.Sprintf("%s.%s", j.Job, j.Env))
}
