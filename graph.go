package reco

import (
	"errors"
	"os"

	"github.com/ReconfigureIO/reco/downloader"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
	humanize "github.com/dustin/go-humanize"
)

// Graph manages reco graphs.
type Graph interface {
	// Generate generates a graph.
	Generate(Args) (output string, err error)
	// List list graphs.
	List(filter M) (printer.Table, error)
	// Open opens a graph.
	Open(id string) (file string, err error)
}

var _ Graph = &platformGraph{}

type platformGraph struct {
	clientImpl
}

func (b platformGraph) prepareGraph() (string, error) {
	projectID, err := b.projectID()
	if err != nil {
		return "", err
	}
	req := b.apiRequest(endpoints.graphs.String())
	reqBody := M{"project_id": projectID}
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

func (p platformGraph) Generate(args Args) (string, error) {
	srcDir := String(args.At(0))
	wait := Bool(args.At(1))

	logger.Info.Println("preparing graph")
	id, err := p.prepareGraph()
	if err != nil {
		return "", err
	}
	logger.Info.Println("done. Graph id: ", id)

	logger.Info.Println("archiving")
	srcArchive, err := archiveDir(srcDir)
	if err != nil {
		return "", err
	}
	logger.Info.Println("done")

	logger.Info.Println("uploading")
	if err := p.uploadJob("graph", id, srcArchive); err != nil {
		return "", err
	}
	logger.Info.Println("done")
	if wait {
		logger.Info.Println("waiting for graph generation to complete")
	}
	logger.Info.Println()

	return id, nil
}

func (p platformGraph) List(filter M) (printer.Table, error) {
	var table printer.Table
	allProjects := filter.Bool("all")
	graphs, err := p.clientImpl.listGraphs(filter)
	if err != nil {
		return table, err
	}
	var body [][]string
	for _, graph := range graphs {
		buildTime := "-"
		if !graph.Time.IsZero() {
			buildTime = humanize.Time(graph.Time)
		}

		row := []string{
			graph.ID,
			graph.Status,
			buildTime,
		}
		if allProjects {
			row = append(row, graph.Project)
		}

		body = append(body, row)
	}

	table = printer.Table{
		Header: []string{"graph id", "status", "requested"},
		Body:   body,
	}
	if allProjects {
		table.Header = append(table.Header, "project")
	}

	return table, nil
}

func (p platformGraph) Open(id string) (string, error) {
	var req = p.apiRequest(endpoints.graphs.Graph())
	req.param("id", id)
	resp, err := req.Do("GET", nil)
	if err != nil {
		return "", err
	}
	switch resp.StatusCode {
	case 404:
		return "", errors.New("graph not found")
	case 204:
		return "", errors.New("no graph generated")
	case 200:
		break
	default:
		return "", errors.New("unknown error occured")
	}

	pdfFile, err := downloader.FromReader(resp.Body, resp.ContentLength)
	if err != nil {
		return "", err
	}

	// rename for easier pdf viewer recognition.
	if err := os.Rename(pdfFile, pdfFile+".pdf"); err == nil {
		pdfFile += ".pdf"
	}

	return pdfFile, nil
}
