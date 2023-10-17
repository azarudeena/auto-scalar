package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
)

func init() {
	// Resty client setup
	client = resty.New()
	client.SetHeader("Accept", "application/json")
	client.SetRetryCount(1)
}

var client *resty.Client

// getAppStatus retrieves the CPU usage and replica count.
// Returns the AppStatus or an error if the retrieval fails.
func getAppStatus() (*AppStatus, error) {
	resp, err := client.R().Get(statusAPIURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("AppStatus failed statusCode: %d, response: %s", resp.StatusCode(), resp.String())
	}

	var status AppStatus
	err = json.Unmarshal(resp.Body(), &status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// calculateReplicaCounts calculates the desired number of replicas based on the current CPU usage.
// It ensures that the number of replicas is adjusted to keep the CPU usage below the target threshold.
func calculateReplicaCounts(status *AppStatus) int {

	currentCPU := status.CPU["highPriority"]
	ratio := currentCPU / targetCPUUsage
	estimate := float64(status.Replicas) * ratio

	if estimate < 1 {
		return 1 // minimum 1 replica is needed.
	}

	return int(estimate + 0.5) // adding 0.5 to round off the estimate.
}

// updateReplicaCount updates the number of replicas for the application.
// Returns an error if the update fails.
func updateReplicaCount(newCount int) error {
	data := Replicas{
		Replicas: newCount,
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Put(replicasAPIURL)

	if err != nil {
		return err
	}

	// Check for 200 OK or 204 No Content
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("update failed, status: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
