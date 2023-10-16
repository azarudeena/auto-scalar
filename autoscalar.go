package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	checkInterval  = 2 * time.Second
)

var (
	client *resty.Client
	rt     = rate.NewLimiter(rate.Every(1*time.Second), 1)
)

func init() {

	// Resty client setup
	client = resty.New()
	client.SetHeader("Accept", "application/json")
	client.SetRetryCount(3)

}

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			monitorAndUpdateReplicas()
			time.Sleep(checkInterval)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")
}

func monitorAndUpdateReplicas() {

	if !rt.Allow() {
		log.Println("Rate Limit Exceeded")
		return
	}

	status, err := getAppStatus()
	if err != nil {
		log.Println("Error fetching app status: %v", err)
		return
	}

	fmt.Println("Retrieved AppStatus CPU :", status.CPU["highPriority"])
	fmt.Println("Retrieved replica count :", status.Replicas)

	newReplicaCounts := calculateReplicaCounts(status)

	fmt.Println("Replica count to update :", newReplicaCounts)

	if newReplicaCounts != status.Replicas {
		err := updateReplicaCount(newReplicaCounts)
		if err != nil {
			log.Println("Error updating replica status: %v", err)
		}
		fmt.Println("Replica count updated to :", newReplicaCounts)
	}
}

// call /app/status to get CPU and replica count. done
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

// calculate the replica new count in a way that CPU <.80 Done
//
//	inc replica will dec CPU and dec replica will inc CPU.
//
// replica is inversely proportional to CPU. calculate replicas as factor of current replicas to target with exiting cpu.
func calculateReplicaCounts(status *AppStatus) int {
	currentCPU := status.CPU["highPriority"]

	ratio := currentCPU / targetCPUUsage

	estimate := float64(status.Replicas) * ratio

	if estimate < 1 {
		return 1
	}

	return int(estimate + 0.5) // adding 0.5 to cover the edge case of 0.8x cpu utils
}

// call /app/replicas to PUT the new replica count - Done
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
