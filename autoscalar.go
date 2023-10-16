package main

import (
	"fmt"
	"time"
)

// declare types for the AppStatus and ReplicaResponse - done
// call /app/status to get CPU and replica count.
// calculate the replica new count in a way that CPU <.80
//		inc replica will dec CPU and dec replica will inc CPU
// call /app/replicas to PUT the new replica count

type AppStatus struct {
	CPU      map[string]float64 `json:"cpu"`
	Replicas int                `json:"replicas"`
}

type Replicas struct {
	Replicas int `json:"replicas"`
}

const (
	targetCPUUsage = 0.80
	statusAPIURL   = "http://localhost:8123/app/status"
	replicasAPIURL = "http://localhost:8123/app/replicas"
	checkInterval  = 10 * time.Second
)

func main() {

	fmt.Println("Idea to Code")
}
