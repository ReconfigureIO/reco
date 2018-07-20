package reco

import (
	"io"
	"time"

	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
	humanize "github.com/dustin/go-humanize"
)

var _ Job = &buildJob{}

type buildJob struct {
	*clientImpl
}

func (b buildJob) prepareBuild(message string) (string, error) {
	projectID, err := b.projectID()
	if err != nil {
		return "", err
	}
	req := b.apiRequest(endpoints.builds.String())
	reqBody := M{"project_id": projectID, "message": message}
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
	err = decodeJSON(resp.Body, &respJSON)
	return respJSON.Value.ID, err
}

func (b buildJob) Start(args Args) (string, error) {
	srcDir := String(args.At(0))
	wait := Bool(args.At(1))
	message := String(args.At(2))

	logger.Info.Println("preparing build")
	id, err := b.prepareBuild(message)
	if err != nil {
		return "", err
	}
	logger.Info.Println("done. Build ID: ", id)

	logger.Info.Println("archiving")
	srcArchive, err := archiveDir(srcDir)
	if err != nil {
		return "", err
	}
	logger.Info.Println("done")

	logger.Info.Println("uploading")
	if err := b.uploadJob("build", id, srcArchive); err != nil {
		return "", err
	}
	logger.Info.Println("done")
	logger.Info.Println()

	if wait {
		b.waitAndLog("build", id)
	}

	return id, nil
}

func (b buildJob) Status(id string) string {
	return b.clientImpl.getStatus("build", id)
}

func (b buildJob) Stop(id string) error {
	return b.clientImpl.stopJob("build", id)
}

func (b buildJob) List(filter M) (printer.Table, error) {
	var table printer.Table
	allProjects := filter.Bool("all")
	builds, err := b.clientImpl.listBuilds(filter)
	if err != nil {
		return table, err
	}
	var body [][]string
	for _, build := range builds {
		buildTime := "-"
		if !build.Time.IsZero() {
			buildTime = humanize.Time(build.Time)
		}

		row := []string{
			build.ID,
			build.Status,
			buildTime,
			timeRounder(build.Duration).Nearest(time.Second),
			build.Message,
		}
		if allProjects {
			row = append(row, build.Project)
		}

		body = append(body, row)
	}

	table = printer.Table{
		Header: []string{"build id", "status", "started", "duration", "message"},
		Body:   body,
	}
	if allProjects {
		table.Header = append(table.Header, "project")
	}

	return table, nil
}

func (b buildJob) Log(id string, writer io.Writer) error {
	return b.clientImpl.logJob("build", id)
}
