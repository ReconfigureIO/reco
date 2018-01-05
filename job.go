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

var (
	// StatusSubmitted is submitted job state.
	StatusSubmitted = "SUBMITTED"
	// StatusQueued is queued job state.
	StatusQueued = "QUEUED"
	// StatusCreatingImage is creating image job state.
	StatusCreatingImage = "CREATING_IMAGE"
	// StatusStarted is started job state.
	StatusStarted = "STARTED"
	// StatusTerminating is terminating job state.
	StatusTerminating = "TERMINATING"
	// StatusTerminated is terminated job state.
	StatusTerminated = "TERMINATED"
	// StatusCompleted is completed job state.
	StatusCompleted = "COMPLETED"
	// StatusErrored is errored job state.
	StatusErrored = "ERRORED"
	// An error event with code value 124 indicates timeout
	StatusTimeout = "TIMED-OUT"
	// An error event with Code value 124 indicates timeout
	ErrorCodeTimeout = 124
)

// jobInfo gives information about a build.
type jobInfo struct {
	ID        string
	Time      time.Time
	Duration  time.Duration
	Status    string
	Project   string
	Command   string
	Build     string
	IPAddress string
}

// UnmarshalJSON customizes JSON decoding for BuildInfo.
func (ji *jobInfo) UnmarshalJSON(b []byte) error {
	var str apiResponse
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if len(str.Job.Events) == 0 {
		str.Job.Events = str.Events
	}
	if str.Project.ID == "" {
		str.Project = str.Build.Project
	}
	ji.ID = str.ID
	ji.Status = "unstarted"
	ji.Project = str.Project.Name
	ji.Command = str.Command
	ji.IPAddress = str.IPAddress
	if str.Build.ID != "" {
		ji.Build = str.Build.ID
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
		for _, event := range str.Job.Events {
			if isTimeout(event) {
				// Timeout isn't a Status reported by API
				// but we know the error event was caused by a timeout
				// so report that to user
				lastEvent.Status = StatusTimeout
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
	case StatusCompleted, StatusErrored, StatusTerminated:
		return true
	}
	return false
}

func (job jobInfo) IsCompleted() bool {
	return isCompleted(job.Status)
}

func (job jobInfo) IsStarted() bool {
	return isStarted(job.Status)
}

// isCompleted checks if the status is a final status.
func isStarted(status string) bool {
	switch strings.ToUpper(status) {
	case StatusCompleted, StatusErrored, StatusTerminated, StatusTerminating, StatusStarted, StatusCreatingImage:
		return true
	}
	return false
}

func isTimeout(ev event) bool {
	if isError(ev) {
		if ev.Code == ErrorCodeTimeout {
			return true
		}
	}
	return false
}

func isError(ev event) bool {
	switch strings.ToUpper(ev.Status) {
	case StatusErrored:
		return true
	}
	return false
}
