package reco

import (
	"errors"
	"fmt"
	"io"
	"net"
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
	Connect(id string, openBrowser bool) error
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

	logger.Info.Println("done. Deployment ID: ", respJSON.Value.ID)
	logger.Info.Println(`you can run "reco deployment log `, respJSON.Value.ID, `" to manually stream logs`)
	if wait == "true" {
		return respJSON.Value.ID, p.waitAndLog("deployment", respJSON.Value.ID)
	} else if wait == "http" {
		err := p.waitForStatus("deployment", respJSON.Value.ID, StatusStarted)
		if err == nil {
			err = p.Connect(respJSON.Value.ID, false)
		}
		return respJSON.Value.ID, err
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

func (p deploymentJob) Connect(id string, openBrowser bool) error {
	logger.Info.Println("Waiting for deployment to listen on port 80")
	for {
		resp, err := p.clientImpl.getJob("deployment", id)
		if err != nil {
			return err
		}

		if resp.IsCompleted() {
			return errors.New("instance has shutdown")
		}

		if resp.IPAddress != "" {
			conn, err := net.DialTimeout("tcp", resp.IPAddress+":80", 5*time.Second)
			if conn != nil && err == nil {
				_ = conn.Close()
				logger.Info.Printf("Deployment ready at http://%s/", resp.IPAddress)
				if openBrowser {
					return open.Run(fmt.Sprintf("http://%s/", resp.IPAddress))
				}
				return nil
			}
		}

		time.Sleep(10 * time.Second)
	}
}
