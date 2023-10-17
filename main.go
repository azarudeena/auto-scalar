package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
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
