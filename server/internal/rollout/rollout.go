package rollout

// InRollout reports whether a device (clientID) gets an update rolled out to
// percent of devices. The same device always gets the same answer.
func InRollout(clientID, groupID string, percent int) bool {
	// TODO: bucket := fnv32a(clientID + ":" + groupID) % 100; return bucket < percent
	return percent >= 100
}
