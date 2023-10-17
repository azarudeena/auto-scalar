package main

// monitorAndUpdateReplicas is a routine that checks the application's CPU usage and updates
// the replica count to ensure the CPU usage stays below the target threshold.
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
