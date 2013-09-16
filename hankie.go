package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
)

type hankie map[Dock][]*Job

func (h *hankie) Register(jobsPath string) error {
	jobs, err := ioutil.ReadDir(jobsPath)
	if err != nil {
		return fmt.Errorf("Couldn't read %s: %s", jobsPath, err)
	}
	for _, jobFile := range jobs {
		if jobFile.IsDir() {
			continue
		}
		jobFilePath := path.Join(jobsPath, jobFile.Name())
		log.Printf("Registering %s", jobFilePath)
		job, err := JobFromFile(jobFilePath)
		if err != nil {
			return fmt.Errorf("Couldn't read job %s: %s", jobFilePath, err)
		}
		h.registerJob(job)
	}
	return nil
}

// - For all jobs in Dock but not in h: Kill em
// - For all jobs in h but not in Dock: Schedule em
func (h *hankie) Converge() error {
	for dock, supposedJobs := range *h {

		currentJobs, err := dock.GetJobs()
		if err != nil {
			return fmt.Errorf("Couldn't get jobs: %s", err)
		}

		log.Printf("[%s] Supposed jobs: %v", dock, supposedJobs)
		log.Printf("[%s]  Current jobs: %v", dock, currentJobs)

		for _, cJob := range currentJobs {
			supposed := false
			for _, sJob := range supposedJobs {
				if cJob.SameAs(sJob) {
					supposed = true
					break
				}
			}
			if !supposed {
				if err := dock.Kill(cJob.Product, cJob.ServiceName()); err != nil {
					return err
				}
			}
		}
		for _, sJob := range supposedJobs {
			scheduled := false
			for _, cJob := range currentJobs {
				if sJob.SameAs(cJob) {
					scheduled = true
					break
				}
			}
			if !scheduled {
				if err := dock.Schedule(sJob); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (h *hankie) registerJob(job *Job) {
	for _, d := range job.Docks {
		dock := Dock(d)
		(*h)[dock] = append((*h)[dock], job)
	}
}
