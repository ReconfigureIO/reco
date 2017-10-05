package reco

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
)

var errNoActiveProject = errors.New("No active project is set, run 'reco project set' to set one")

var _ ProjectConfig = &platformProject{}

// ProjectConfig manages projects.
type ProjectConfig interface {
	// List lists the projects.
	List(filter M) (printer.Table, error)
	// list lists information of projects.
	list() ([]ProjectInfo, error)
	// Create create a new project.
	Create(name string) error
	// Set sets the active project.
	Set(name string) error
	// Get gets the name of the active project.
	Get() (string, error)
}

// ProjectInfo gives information about a project.
type ProjectInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"-"`
}

type platformProject struct {
	p clientImpl
}

func (p platformProject) List(filter M) (printer.Table, error) {
	var table printer.Table
	req := p.p.apiRequest(endpoints.projects.String())
	resp, err := req.Do("GET", nil)
	if err != nil {
		return table, err
	}
	var jsonResp struct {
		Value []ProjectInfo `json:"value"`
		Error string        `json:"error"`
	}
	err = decodeJSON(resp.Body, &jsonResp)
	if err != nil {
		return table, err
	}

	active := false
	// TODO remove set when platform handles distinct names
	set := make(map[string]struct{})
	var body [][]string
	for _, v := range jsonResp.Value {
		set[v.Name] = struct{}{}
		row := []string{v.Name}

		// active
		if v.ID == p.p.ProjectID {
			row = append(row, "[*]")
			active = true
		}
		body = append(body, row)
	}

	table = printer.Table{
		Header: []string{"name", "active"},
		Body:   body,
	}
	if !active {
		logger.Std.Println(errNoActiveProject)
	}
	return table, nil
}

func (p platformProject) Create(name string) error {
	req := p.p.apiRequest(endpoints.projects.String())
	reqBody := M{"name": name}
	resp, err := req.Do("POST", reqBody)
	if err != nil {
		return err
	}
	var jsonResp struct {
		Value ProjectInfo `json:"value"`
		Error string      `json:"error"`
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.New("project not created")
	}
	if err := decodeJSON(resp.Body, &jsonResp); err != nil {
		return err
	}
	// if no project is set, attempt to use this one
	if p.p.ProjectID == "" {
		p.p.ProjectID = jsonResp.Value.ID
		if err := p.p.saveProject(); err != nil {
			logger.Error.Println(err)
		}
	}
	return nil
}

func (p platformProject) Set(name string) error {
	projects, err := p.list()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		return errProjectNotCreated
	}
	id := ""
	for i := range projects {
		if strings.ToLower(projects[i].Name) == strings.ToLower(name) {
			id = projects[i].ID
			break
		}
	}
	if id == "" {
		return errProjectNotFound
	}
	p.p.ProjectID = id
	return p.p.saveProject()
}

func (p platformProject) list() ([]ProjectInfo, error) {
	req := p.p.apiRequest(endpoints.projects.String())
	resp, err := req.Do("GET", nil)
	if err != nil {
		return nil, err
	}
	var jsonResp struct {
		Value []ProjectInfo `json:"value"`
		Error string        `json:"error"`
	}
	err = decodeJSON(resp.Body, &jsonResp)
	if err != nil {
		return nil, err
	}
	// set active project
	for i, v := range jsonResp.Value {
		if v.ID == p.p.ProjectID {
			jsonResp.Value[i].Active = true
			break
		}
	}
	return jsonResp.Value, err
}

func (p platformProject) Get() (string, error) {
	projects, err := p.list()
	if err != nil {
		return "", err
	}
	for _, prj := range projects {
		if prj.Active {
			return prj.Name, nil
		}
	}
	return "", errNoActiveProject
}
