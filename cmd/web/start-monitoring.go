package main

type job struct {
	HostServiceID int
}

func (j job) Run() {

}

func startMonitoring() {
	if preferenceMap["monitoring_live"] == "1" {
		data := make(map[string]string)
		data["message"] = "starting"

		// trigger a message to broadcast to all clients that app is starting to monitor

		// get all of the services that needs to be monitor

		// create a job for every service and broadcast over websockets

	}
}
