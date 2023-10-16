package main

// call /app/replicas to PUT the new replica count

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"time"
)

// declare types for the AppStatus and Replicas - done
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
	checkInterval  = 5 * time.Second
)

var client *resty.Client

func init() {

	// Resty client setup
	client = resty.New()
	client.SetHeader("Accept", "application/json")
	client.SetRetryCount(3)

}

func main() {

	status, err := getAppStatus()
	if err != nil {
		log.Fatalf("Error fetching app status: %v", err)
	}

	fmt.Println("Retrieved AppStatus CPU :", status.CPU["highPriority"])
	fmt.Println("Retrieved replica count :", status.Replicas)

	newReplicaCounts := calculateReplicaCounts(status)

	fmt.Println("Replica count to update :", newReplicaCounts)
}

// call /app/status to get CPU and replica count. done
func getAppStatus() (*AppStatus, error) {
	resp, err := client.R().Get(statusAPIURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("AppStatus statusCode: %d, response: %s", resp.StatusCode(), resp.String())
	}

	var status AppStatus
	err = json.Unmarshal(resp.Body(), &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// calculate the replica new count in a way that CPU <.80 Done
//
//	inc replica will dec CPU and dec replica will inc CPU.
//
// replica is inversely proportional to CPU. factor of current cpu to target with exiting cpu.
func calculateReplicaCounts(status *AppStatus) int {
	currentCPU := status.CPU["highPriority"]

	ratio := currentCPU / targetCPUUsage

	estimate := float64(status.Replicas) * ratio

	if estimate < 1 {
		return 1
	}

	return int(estimate + 0.5) // adding 0.5 to cover the edge case of 0.8x cpu utils
}
