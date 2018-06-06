package reco

import (
	"testing"
)

var (
	jobsComplete = []jobInfo{
		jobInfo{Status: "terminated"},
		jobInfo{Status: "errored"},
		jobInfo{Status: "completed"},
		jobInfo{Status: "TERMINATED"},
		jobInfo{Status: "ERRORED"},
		jobInfo{Status: "COMPLETED"},
	}

	jobsNotComplete = []jobInfo{
		jobInfo{Status: "submitted"},
		jobInfo{Status: "queued"},
		jobInfo{Status: "creating_image"},
		jobInfo{Status: "started"},
		jobInfo{Status: "terminating"},
		jobInfo{Status: "SUBMITTED"},
		jobInfo{Status: "QUEUED"},
		jobInfo{Status: "CREATING_IMAGE"},
		jobInfo{Status: "STARTED"},
		jobInfo{Status: "TERMINATING"},
	}

	jobsStarted = []jobInfo{
		jobInfo{Status: "started"},
		jobInfo{Status: "terminating"},
		jobInfo{Status: "creating_image"},
		jobInfo{Status: "terminated"},
		jobInfo{Status: "errored"},
		jobInfo{Status: "completed"},
		jobInfo{Status: "TERMINATED"},
		jobInfo{Status: "ERRORED"},
		jobInfo{Status: "COMPLETED"},
		jobInfo{Status: "CREATING_IMAGE"},
		jobInfo{Status: "STARTED"},
		jobInfo{Status: "TERMINATING"},
	}

	jobsNotStarted = []jobInfo{
		jobInfo{Status: "submitted"},
		jobInfo{Status: "queued"},
		jobInfo{Status: "SUBMITTED"},
		jobInfo{Status: "QUEUED"},
	}
)

func TestIsCompleted(t *testing.T) {
	for _, job := range jobsComplete {
		if !job.IsCompleted() {
			t.Error("IsCompleted returned false when it should be true")
		}
	}
	for _, job := range jobsNotComplete {
		if job.IsCompleted() {
			t.Error("IsCompleted returned true when it should be false")
		}
	}
}

func TestIsStarted(t *testing.T) {
	for _, job := range jobsStarted {
		if !job.IsStarted() {
			t.Error("IsStarted returned false when it should be true")
		}
	}
	for _, job := range jobsNotStarted {
		if job.IsStarted() {
			t.Error("IsStarted returned true when it should be false")
		}
	}
}
