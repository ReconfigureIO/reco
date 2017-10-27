package reco

import (
	"encoding/json"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ReconfigureIO/reco/printer"
)

// Job is a set of actions for the reco platform.
type Job interface {
	// Start starts the job.
	Start(Args) (output string, err error)
	// Stop stops the job.
	Stop(id string) error
	// List lists job resources.
	List(filter M) (printer.Table, error)
	// Log logs the job.
	Log(id string, writer io.Writer) error
}

// jobInfo gives information about a build.
type jobInfo struct {
	ID       string
	Time     time.Time
	Duration time.Duration
	Status   string
	Project  string
	Command  string
	Build    string
}

// UnmarshalJSON customizes JSON decoding for BuildInfo.
func (ji *jobInfo) UnmarshalJSON(b []byte) error {
	var str apiResponse
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	ji.ID = str.ID
	ji.Status = "unstarted"
	ji.Project = str.Project.Name
	ji.Command = str.Command
	if str.Build.ID != "" {
		ji.Build = str.Build.ID
	}
	if len(str.Job.Events) == 0 {
		str.Job.Events = str.Events
	}
	sort.Sort(eventSorter(str.Job.Events))
	if len(str.Job.Events) > 0 {
		firstEvent := str.Job.Events[0]
		lastEvent := str.Job.Events[len(str.Job.Events)-1]
		// Handle terminated status with a prior final status.
		if len(str.Job.Events) > 2 {
			ev := str.Job.Events[len(str.Job.Events)-2]
			if isCompleted(ev.Status) {
				lastEvent = ev
			}
		}
		ji.Status = strings.ToLower(lastEvent.Status)
		ji.Time = firstEvent.Timestamp
		if eventSorter(str.Job.Events).Completed() {
			ji.Duration = lastEvent.Timestamp.Sub(firstEvent.Timestamp)
		}
	}

	return nil
}

type jobSorter []jobInfo

func (b jobSorter) Len() int           { return len(b) }
func (b jobSorter) Less(i, j int) bool { return b[i].Time.After(b[j].Time) }
func (b jobSorter) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

type jobFilter []jobInfo

func (b jobFilter) doFilter(f func(*jobInfo) bool) (bi []jobInfo) {
	for i := range b {
		info := b[i]
		if f(&info) {
			bi = append(bi, info)
		}
	}
	return
}

func (b jobFilter) Filter(key, val string) (bi []jobInfo) {
	switch key {
	case "status":
		return b.doFilter(func(info *jobInfo) bool {
			return strings.ToUpper(info.Status) == strings.ToUpper(val)
		})
	case "id":
		return b.doFilter(func(info *jobInfo) bool {
			return info.ID == val
		})
	}
	return
}

type eventSorter []event

func (e eventSorter) Less(i, j int) bool { return e[i].Timestamp.Before(e[j].Timestamp) }
func (e eventSorter) Len() int           { return len(e) }
func (e eventSorter) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e eventSorter) Completed() bool {
	if len(e) < 2 {
		return false
	}
	return isCompleted(e[len(e)-1].Status)
}

// isCompleted checks if the status is a final status.
func isCompleted(status string) bool {
	switch strings.ToUpper(status) {
	case "COMPLETED", "ERRORED", "TERMINATED":
		return true
	}
	return false
}
