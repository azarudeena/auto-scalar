package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"cpu": {"highPriority": 0.5}, "replicas": 2}`)
	}))
	defer ts.Close()

	statusAPIURL = ts.URL

	status, err := getAppStatus()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status.CPU["highPriority"] != 0.5 {
		t.Errorf("Expected CPU highPriority to be 0.5, got %f", status.CPU["highPriority"])
	}
	if status.Replicas != 2 {
		t.Errorf("Expected replicas to be 2, got %d", status.Replicas)
	}
}

func TestGetAppStatus_ErrorScenario(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal server error")
	}))
	defer ts.Close()

	statusAPIURL = ts.URL

	_, err := getAppStatus()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func TestCalculateNewReplicaCount(t *testing.T) {
	status := &AppStatus{
		CPU: map[string]float64{
			"highPriority": 0.70,
		},
		Replicas: 10,
	}

	newReplicaCount := calculateReplicaCounts(status)
	assert.Equal(t, 9, newReplicaCount) // Based on the logic, this should be the result.
}

func TestCalculateNewReplicaCount_Maintain1Replica(t *testing.T) {
	status := &AppStatus{
		CPU: map[string]float64{
			"highPriority": 0.1,
		},
		Replicas: 1,
	}

	newReplicaCount := calculateReplicaCounts(status)
	assert.Equal(t, 1, newReplicaCount) // Based on the logic, this should be the result.
}

func TestUpdateReplicaCount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected method PUT, got %s", r.Method)
		}
	}))
	defer ts.Close()

	replicasAPIURL = ts.URL

	err := updateReplicaCount(5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUpdateReplicaCount_ErrorScenario(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal server error")
	}))
	defer ts.Close()

	replicasAPIURL = ts.URL

	err := updateReplicaCount(5)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}
