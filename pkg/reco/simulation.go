package reco

import (
	"io"
	"strings"
	"time"

	"github.com/ReconfigureIO/reco/pkg/logger"
	"github.com/ReconfigureIO/reco/pkg/printer"
	humanize "github.com/dustin/go-humanize"
)

var _ Job = &testJob{}

type testJob struct {
	*clientImpl
}

func (t testJob) prepareTest(command string) (string, error) {
	projectID, err := t.projectID()
	if err != nil {
		return "", err
	}
	req := t.apiRequest(endpoints.simulations.String())
	reqBody := M{"project_id": projectID, "command": command}
	resp, err := req.Do("POST", reqBody)
	if err != nil {
		return "", err
	}
	var respJSON struct {
		Value struct {
			ID string `json:"id"`
		} `json:"value"`
		Error string `json:"error,omitempty"`
	}
	err = decodeJSON(resp.Body, &respJSON)
	return respJSON.Value.ID, err
}

func (p testJob) Start(args Args) (string, error) {
	srcDir := String(args.At(0))
	cmd := String(args.At(1))
	cmdArgs := StringSlice(args.At(2))

	if len(cmdArgs) > 0 {
		cmd += " " + strings.Join(cmdArgs, " ")
	}
	logger.Info.Println("preparing simulation")
	id, err := p.prepareTest(cmd)
	if err != nil {
		return "", err
	}
	logger.Info.Println("done")

	logger.Info.Println("archiving")
	srcArchive, err := archiveDir(srcDir)
	if err != nil {
		return "", err
	}
	logger.Info.Println("done")

	logger.Info.Println("uploading")
	if err := p.uploadJob("simulation", id, srcArchive); err != nil {
		return "", err
	}
	logger.Info.Println("done")

	logger.Info.Println("running simulation")
	logger.Info.Println()
	p.waitAndLog("simulation", id)

	return id, nil
}

func (b testJob) List(filter M) (printer.Table, error) {
	var table printer.Table
	allProjects := filter.Bool("all")
	simulations, err := b.clientImpl.listTests(filter)
	if err != nil {
		return table, err
	}
	var body [][]string
	for _, sim := range simulations {
		simTime := "-"
		if !sim.Time.IsZero() {
			simTime = humanize.Time(sim.Time)
		}

		row := []string{
			sim.ID,
			sim.Status,
			simTime,
			timeRounder(sim.Duration).Nearest(time.Second),
		}
		if allProjects {
			row = append(row, sim.Project)
		}

		body = append(body, row)
	}

	table = printer.Table{
		Header: []string{"simulation id", "status", "started", "duration"},
		Body:   body,
	}
	if allProjects {
		table.Header = append(table.Header, "project")
	}

	return table, nil
}

func (t testJob) Status(id string) string {
	return t.clientImpl.getStatus("simulation", id)
}

func (t testJob) Stop(id string) error {
	return t.clientImpl.stopJob("simulation", id)
}

func (t testJob) Log(id string, writer io.Writer) error {
	return t.clientImpl.logJob("simulation", id)
}
