package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/tsawler/vigilate/internal/helpers"
	"github.com/tsawler/vigilate/internal/models"

	"github.com/CloudyKit/jet/v6"
)

type ByHost []models.Schedule

// Len returns the length of the ByHost slice
func (a ByHost) Len() int { return len(a) }

// Less returns true if the item at index i is less than the item at index j
func (a ByHost) Less(i, j int) bool { return a[i].Host < a[j].Host }

// Swap swaps the items at index i and j
func (a ByHost) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	for k, v := range repo.App.MonitorMap {
		var item models.Schedule

		item.ID = k
		item.EntryID = v
		item.Entry = repo.App.Scheduler.Entry(v)
		hs, err := repo.DB.GetHostServiceByID(k)
		if err != nil {
			printTemplateError(w, err)
			return
		}
		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
		item.LastRunFromHS = hs.LastCheck
		item.Host = hs.HostName
		item.Service = hs.Service.ServiceName
		items = append(items, item)
	}

	// sort the items by host
	sort.Sort(ByHost(items))

	data := make(jet.VarMap)
	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
