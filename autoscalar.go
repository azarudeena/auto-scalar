package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

const ()

var (
	client         *resty.Client
	logger         = logrus.New()
	targetCPUUsage float64
	statusAPIURL   string
	environment    string
	replicasAPIURL string
	checkInterval  time.Duration
)

func init() {

	// Config Setup
	viper.SetDefault("STATUS_API_URL", "http://localhost:8123/app/status")
	viper.SetDefault("REPLICAS_API_URL", "http://localhost:8123/app/replicas")
	viper.SetDefault("CHECK_INTERVAL", "5s")
	viper.SetDefault("TARGET_CPU_USAGE", 0.80)
	viper.SetDefault("ENVIRONMENT", "dev")

	viper.AutomaticEnv()

	statusAPIURL = viper.GetString("STATUS_API_URL")
	replicasAPIURL = viper.GetString("REPLICAS_API_URL")
	checkInterval = viper.GetDuration("CHECK_INTERVAL")
	targetCPUUsage = viper.GetFloat64("TARGET_CPU_USAGE")
	environment = viper.GetString("ENVIRONMENT")

	// Logger setup
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	if environment == "prod" {
		logger.SetLevel(logrus.InfoLevel)
	}

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
	logger.Info("Shutting down gracefully...")
}

func monitorAndUpdateReplicas() {

	status, err := getAppStatus()
	if err != nil {
		logger.Errorf("Error fetching app status: %v", err)
		return
	}

	logger.Debugf("Retrieved AppStatus CPU : %f ", status.CPU["highPriority"])
	logger.Debugf("Retrieved replica count: %d", status.Replicas)

	newReplicaCounts := calculateReplicaCounts(status)

	logger.Debugf("Replica count to update : %d", newReplicaCounts)

	if newReplicaCounts != status.Replicas {
		err := updateReplicaCount(newReplicaCounts)
		if err != nil {
			logger.Errorf("Error updating replica status: %v", err)
		}
		logger.Infof("Replica count updated to :%d", newReplicaCounts)
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
		return 1 // minimum 1 replica is needed.
	}

	return int(estimate + 0.5) // adding 0.5 to cover the edge case of 0.8x cpu utils
}

// call /app/replicas to PUT the new replica count
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
