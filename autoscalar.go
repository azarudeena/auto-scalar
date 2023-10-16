package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"time"
)

// declare types for the AppStatus and ReplicaResponse - done
// call /app/status to get CPU and replica count. done
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
}

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
