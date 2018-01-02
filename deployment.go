package reco

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
	humanize "github.com/dustin/go-humanize"
	"github.com/skratchdot/open-golang/open"
)

// DeploymentProxy proxies to a running deployment instance.
type DeploymentProxy interface {
	// Connect performs a proxy connection.
	Connect(id string) error
}

var _ Job = deploymentJob{}
var _ DeploymentProxy = deploymentJob{}

type deploymentJob struct {
	clientImpl
}

func (p deploymentJob) Start(args Args) (string, error) {
	buildID := String(args.At(0))
	command := String(args.At(1))
	wait := String(args.Last())
	cmdArgs := StringSlice(args.At(2))

	req := p.apiRequest(endpoints.deployments.String())
	if len(args) > 0 {
		command += " " + strings.Join(cmdArgs, " ")
	}

	logger.Info.Println("creating deployment")
	reqBody := M{
		"build_id": buildID,
		"command":  command,
	}
	resp, err := req.Do("POST", reqBody)
	if err != nil {
		return "", err
	}
	var respJSON struct {
		Value struct {
			ID string `json:"id"`
		} `json:"value"`
		Error string `json:"error"`
	}
	if err := decodeJSON(resp.Body, &respJSON); err != nil {
		return "", err
	}
	if respJSON.Error != "" {
		return "", errors.New(respJSON.Error)
	}
	if respJSON.Value.ID == "" {
		return "", errUnknownError
	}

	logger.Info.Println("done. Deployment id: ", respJSON.Value.ID)
	logger.Info.Println()
	if wait == "true" {
		return respJSON.Value.ID, p.waitAndLog("deployment", respJSON.Value.ID)
	} else if wait == "http" {
		return respJSON.Value.ID, p.Connect(respJSON.Value.ID)
	}
	return "", nil
}

func (p deploymentJob) Stop(id string) error {
	resp, err := p.clientImpl.getJob("deployment", id)
	if err != nil {
		return err
	}
	if !resp.IsCompleted() {
		return p.clientImpl.stopJob("deployment", id)
	} else {
		return nil
	}
}

func (p deploymentJob) List(filter M) (printer.Table, error) {
	var table printer.Table
	allProjects := filter.Bool("all")
	deployments, err := p.clientImpl.listDeployments(filter)
	if err != nil {
		return table, err
	}

	var body [][]string
	for _, deployment := range deployments {
		buildTime := "-"
		if !deployment.Time.IsZero() {
			buildTime = humanize.Time(deployment.Time)
		}

		row := []string{
			deployment.ID,
			deployment.Build,
			deployment.Command,
			deployment.Status,
			buildTime,
			timeRounder(deployment.Duration).Nearest(time.Second),
		}
		if allProjects {
			row = append(row, deployment.Project)
		}

		body = append(body, row)
	}

	table = printer.Table{
		Header: []string{"id", "image", "command", "status", "started", "duration"},
		Body:   body,
	}
	if allProjects {
		table.Header = append(table.Header, "project")
	}

	return table, nil
}

func (p deploymentJob) Log(id string, writer io.Writer) error {
	return p.clientImpl.logJob("deployment", id)
}

func (p deploymentJob) Connect(id string) error {
	for {
		resp, err := p.clientImpl.getJob("deployment", id)
		if err != nil {
			return err
		}

		if resp.IsCompleted() {
			return errors.New("instance has shutdown")
		}

		if resp.IPAddress != "" {
			return open.Run(fmt.Sprintf("http://%s/", resp.IPAddress))
		}

		time.Sleep(10 * time.Second)
	}
}
