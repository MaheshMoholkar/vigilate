package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type job struct {
	HostServiceID int
}

func (j *job) Run() {
	Repo.ScheduledCheck(j.HostServiceID)
}

// StartMonitoring starts the monitoring of the services
func (repo *DBRepo) StartMonitoring() {
	if app.PreferenceMap["monitoring_live"] == "1" {
		data := make(map[string]string)
		data["message"] = "Monitoring is starting..."

		err := app.WsClient.Trigger("public-channel", "app-starting", data)
		if err != nil {
			log.Println(err)
		}

		// get all of the services that are being monitored
		servicesToMonitor, err := repo.DB.GetServicesToMonitor()
		if err != nil {
			log.Println(err)
		}

		// loop through the services
		for _, x := range servicesToMonitor {
			// get the schedule unit number
			var sch string
			if x.ScheduleUnit == "d" {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber*24, "h")
			} else {
				sch = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)
			}

			// create a job
			var j job
			j.HostServiceID = x.ID
			scheduleID, err := app.Scheduler.AddJob(sch, &j)
			if err != nil {
				log.Println(err)
			}

			// save id job so start/stop it
			app.MonitorMap[x.ID] = scheduleID

			// broadcast over websockets the fact that the service is scheduled
			payload := make(map[string]string)
			payload["message"] = "scheduling"
			payload["host_service_id"] = strconv.Itoa(x.ID)
			yearOne := time.Date(0001, 11, 17, 20, 34, 58, 65138737, time.UTC)

			if app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.After(yearOne) {
				data["next_run"] = app.Scheduler.Entry(app.MonitorMap[x.ID]).Next.Format("2006-01-02 3:04:05 PM")
			} else {
				data["next_run"] = "Pending..."
			}

			payload["host"] = x.HostName
			payload["service"] = x.Service.ServiceName
			if x.LastCheck.After(yearOne) {
				payload["last_run"] = x.LastCheck.Format("2006-01-02 3:04:05 PM")
			} else {
				payload["last_run"] = "Pending..."
			}

			payload["schedule"] = fmt.Sprintf("@every %d%s", x.ScheduleNumber, x.ScheduleUnit)

			err = app.WsClient.Trigger("public-channel", "next-run-event", payload)
			if err != nil {
				log.Println(err)
			}

			err = app.WsClient.Trigger("public-channel", "schedule-changed-event", payload)
			if err != nil {
				log.Println(err)
			}
		}

	}
}
