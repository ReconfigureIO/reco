package reco

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/ReconfigureIO/reco/logger"
	"github.com/spf13/viper"
)

const (
	// GlobalConfigDirKey is the key for reco configuration in viper.
	GlobalConfigDirKey = "reco_global_config_dir"
	// ConfigDirKey is the key for reco configuration for current directory.
	ConfigDirKey = "reco_config_dir"

	// waitInterval is the interval to wait before checking build status updates.
	waitInterval = time.Second * 5

	platformServerKey     = "PLATFORM_SERVER"
	platformServerAddress = "https://api.reconfigure.io"
	platformAuthFile      = "auth.json"
	platformProjectFile   = "project.json"
)

var (
	errUnsupported       = errors.New("command is unsupported for reconfigure.io platform")
	errMissingServer     = errors.New("PLATFORM_SERVER config or environment variable not set")
	errAuthRequired      = errors.New("authentication required. Run 'reco auth' to authenticate")
	errAuthFailed        = errors.New("authentication failed. Run 'reco auth' to authenticate")
	errProjectNotSet     = errors.New("project not set. Run 'reco project set' to set one")
	errProjectNotCreated = errors.New("no projects found. Run 'reco project create' to create one")
	errProjectNotFound   = errors.New("project not found")
	errNetworkError      = errors.New("network error")
	errNotFound          = errors.New("not found")
	errInvalidToken      = errors.New("the token is invalid")
	errUnknownError      = errors.New("unknown error occurred")
	errBadResponse       = errors.New("bad response from server")

	// StatusSubmitted is submitted job state.
	StatusSubmitted = []string{"SUBMITTED", "submitted"}
	// StatusQueued is queued job state.
	StatusQueued = []string{"QUEUED", "queued"}
	// StatusCreatingImage is creating image job state.
	StatusCreatingImage = []string{"CREATING_IMAGE", "creating_image"}
	// StatusStarted is started job state.
	StatusStarted = []string{"STARTED", "started"}
	// StatusTerminating is terminating job state.
	StatusTerminating = []string{"TERMINATING", "terminating"}
	// StatusTerminated is terminated job state.
	StatusTerminated = []string{"TERMINATED", "terminated"}
	// StatusCompleted is completed job state.
	StatusCompleted = []string{"COMPLETED", "completed"}
	// StatusErrored is errored job state.
	StatusErrored = []string{"ERRORED", "errored"}
	// StatusWaiting is waiting for events state
	StatusWaiting     = "Waiting"
	JobTypeBuild      = "build"
	JobTypeDeployment = "deployment"
	JobTypeSimulation = "simulation"
	JobTypeGraph      = "graph"
)

// Client is a reconfigure.io platform client.
type Client interface {
	// Init initiates the client.
	Init() error
	// Auth authenticates the user.
	Auth(token string) error
	// Test handles simulation actions.
	Test() Job
	// Build handles build actions.
	Build() Job
	// Deployment handles deployment actions.
	Deployment() Job
	// Project handles project actions.
	Project() ProjectConfig
	// Graph handles graph actions.
	Graph() Graph
}

var _ Client = &clientImpl{}

// clientImpl is the implementation of reconfigure.io client
type clientImpl struct {
	platformServer string
	Username       string `json:"user_id,omitempty"`
	Token          string `json:"token,omitempty"`
	ProjectID      string `json:"project,omitempty"`
}

// NewClient creates a new reconfigure.io client.
func NewClient() Client {
	return &clientImpl{}
}

func (p clientImpl) Build() Job {
	return &buildJob{p}
}
func (p clientImpl) Test() Job {
	return &testJob{p}
}
func (p clientImpl) Deployment() Job {
	return &deploymentJob{p}
}
func (p clientImpl) Project() ProjectConfig {
	return &platformProject{p}
}
func (p clientImpl) Graph() Graph {
	return &platformGraph{p}
}

func (p *clientImpl) initAuth() {
	if p.Username == "" && p.Token == "" {
		p.loadAuth()
	}
}

func (p *clientImpl) initProject() {
	if p.ProjectID == "" {
		p.loadProject()
	}
}

func (p clientImpl) authFileName() string {
	return filepath.Join(viper.GetString(GlobalConfigDirKey), platformAuthFile)
}

func (p clientImpl) projectFileName() string {
	return filepath.Join(viper.GetString(ConfigDirKey), platformProjectFile)
}

func (p clientImpl) saveAuth() error {
	var prj clientImpl
	if p.Username == "" || p.Token == "" {
		return nil
	}
	prj.Username = p.Username
	prj.Token = p.Token
	authFile, err := os.OpenFile(p.authFileName(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer authFile.Close()
	return json.NewEncoder(authFile).Encode(prj)
}

func (p clientImpl) saveProject() error {
	var prj clientImpl
	if p.ProjectID == "" {
		return nil
	}
	prj.ProjectID = p.ProjectID
	projectFile, err := os.OpenFile(p.projectFileName(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer projectFile.Close()
	return json.NewEncoder(projectFile).Encode(prj)
}

func (p *clientImpl) loadAuth() error {
	f, err := os.Open(p.authFileName())
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(p)
}

func (p *clientImpl) loadProject() error {
	f, err := os.Open(p.projectFileName())
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(p)
}

func (p *clientImpl) Init() error {
	server := viper.GetString(platformServerKey)
	if server == "" {
		server = platformServerAddress
	}
	u, err := url.Parse(server)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	p.platformServer = u.String()
	p.initAuth()
	p.initProject()
	return nil
}

func (p *clientImpl) Auth(token string) error {
	authURL := fmt.Sprint("http://app.reconfigure.io/dashboard")
	if token == "" {
		fmt.Println("Copy the token after authentication at", authURL)
		fmt.Print("Token: ")
		if _, err := fmt.Scanln(&token); err != nil {
			return err
		}
	}
	str := strings.Split(token, "_")
	if len(str) != 3 {
		return errInvalidToken
	}
	p.Username = str[0] + "_" + str[1]
	p.Token = str[2]
	return p.saveAuth()
}

func jsonToReader(j interface{}) (io.Reader, error) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(j)
	return &b, err
}

func decodeJSON(r io.Reader, body interface{}) error {
	if err := json.NewDecoder(r).Decode(body); err != nil {
		return errBadResponse
	}
	// Check if decode struct has Error field.
	// Return error if field is present and not empty.
	if reflect.ValueOf(body).Kind() != reflect.Ptr {
		return nil
	}
	errField := reflect.ValueOf(body).Elem().FieldByName("Error")
	if e, ok := errField.Interface().(string); ok && e != "" {
		return errors.New(e)
	}
	return nil
}

func (p clientImpl) logJob(eventType string, id string) error {
	logger.Info.Println("streaming logs for ", eventType, " ", id)
	return p.logs(eventType, id)
}

func (p clientImpl) waitAndLog(jobType string, id string) error {
	logger.Info.Println(`you can run "reco `, jobType, " log ", id, `" to manually stream logs`)
	logger.Info.Println("getting ", jobType, " details")
	getStatus := func() string {
		job, err := p.getJob(jobType, id)
		if err == nil && job.Status != "" {
			return job.Status
		}
		return StatusWaiting
	}

	status := getStatus()
	logger.Info.Println("status: ", status)

	if jobType == "" || jobType == JobTypeBuild {
		logger.Info.Println("this will take at least 4 hours")
	} else {
		logger.Info.Println("this may take several minutes")
	}

	logger.Info.Println("waiting for ", jobType, " to start")

	body, err := p.waitForLog(jobType, id, true)
	if err != nil {
		return err
	}
	defer body.Close()

	logger.Info.Println("status: ", getStatus())
	logger.Info.Println("streaming logs")
	logger.Std.Println()
	return p.logs(jobType, id)
}

func (p clientImpl) logs(jobType string, id string) error {
	_, err := p.waitForLog(jobType, id, false)
	return err
}

func (p clientImpl) getJob(jobType string, id string) (jobInfo, error) {
	var apiResp struct {
		Job jobInfo `json:"value"`
	}
	var endpoint string
	switch jobType {
	case JobTypeSimulation:
		endpoint = endpoints.simulations.Item()
	case JobTypeDeployment:
		endpoint = endpoints.deployments.Item()
	default:
		endpoint = endpoints.builds.Item()
	}
	req := p.apiRequest(endpoint)
	req.param("id", id)
	resp, err := req.Do("GET", nil)
	if err != nil {
		return apiResp.Job, err
	}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)
	return apiResp.Job, err
}

// waitForLog attempts to stream logs. If peek is true, it ensures log streaming has started
// and returns the body for the caller to read remaining contents.
// Otherwise, logs are streamed to stderr.
func (p clientImpl) waitForLog(jobType, id string, peek bool) (io.ReadCloser, error) {
	var endpoint string
	switch jobType {
	case JobTypeSimulation:
		endpoint = endpoints.simulations.Log()
	case JobTypeDeployment:
		endpoint = endpoints.deployments.Log()
	default:
		endpoint = endpoints.builds.Log()
	}
	req := p.apiRequest(endpoint)
	req.param("id", id)

	resp, err := req.Do("GET", nil)
	if err != nil {
		return nil, err
	}

	if peek {
		// just verify that the server is streaming response
		// and pass over the remaining body
		var buf = make([]byte, 1)
		var err error
		for {
			_, err = resp.Body.Read(buf)
			if err != nil {
				break
			}
			if buf[0] != 0 {
				os.Stderr.Write(buf)
				break
			}
		}
		return resp.Body, err
	} else {
		defer resp.Body.Close()
		_, err = io.Copy(os.Stderr, resp.Body)
		return nil, err
	}
}

func (p clientImpl) uploadJob(jobType string, id string, srcArchive string) error {
	var endpoint string
	switch jobType {
	case JobTypeSimulation:
		endpoint = endpoints.simulations.Input()
	case JobTypeGraph:
		endpoint = endpoints.graphs.Input()
	default:
		endpoint = endpoints.builds.Input()
	}
	req := p.apiRequest(endpoint)
	req.param("id", id)
	req.jsonBody = false

	f, err := os.Open(srcArchive)
	if err != nil {
		return err
	}

	resp, err := req.Do("PUT", f)
	var respJSON struct {
		Value apiResponse `json:"value"`
		Error string      `json:"error"`
	}

	decodeJSON(resp.Body, &respJSON)

	if resp.StatusCode > 299 || len(respJSON.Value.Job.Events) == 0 {
		return errors.New("unknown error occured")
	}
	return err
}

func (p clientImpl) apiRequest(endpoint string) clientRequest {
	return clientRequest{
		endpoint: p.platformServer + endpoint,
		username: p.Username,
		password: p.Token,
		jsonBody: true,
	}
}

func (p clientImpl) projectID() (string, error) {
	if p.Username == "" || p.Token == "" {
		return "", errAuthRequired
	}
	var prjs []ProjectInfo
	prjs, err := p.Project().list()
	switch err {
	case errAuthFailed, errNetworkError:
		return "", err
	}

	// check if project flag is set
	if prjName := viper.GetString("project"); prjName != "" {
		// extract project ID
		for _, prj := range prjs {
			if prj.Name == prjName {
				return prj.ID, nil
			}
		}
		return "", errProjectNotFound
	}

	if p.ProjectID == "" {
		if len(prjs) == 0 {
			return "", errProjectNotCreated
		}
		return "", errProjectNotSet
	}

	// return configured project
	return p.ProjectID, nil
}

func (p clientImpl) listJobs(jobType string, filters M) ([]jobInfo, error) {
	limit := filters.Int("limit")

	var endpoint string
	switch jobType {
	case JobTypeDeployment:
		endpoint = endpoints.deployments.String()
	case JobTypeSimulation, "test":
		endpoint = endpoints.simulations.String()
	case JobTypeGraph:
		endpoint = endpoints.graphs.String()
	default:
		endpoint = endpoints.builds.String()
	}

	request := p.apiRequest(endpoint)

	// if all-projects flag is not set,
	// and public flag not set, use specific project.
	if !filters.Bool("all") && !filters.Bool("public") {
		projectID, err := p.projectID()
		if err != nil {
			return nil, err
		}
		request.queryParam("project", projectID)
	}

	if filters.Bool("public") {
		request.queryParam("public", "true")
	}

	resp, err := request.Do("GET", nil)
	if err != nil {
		return nil, err
	}
	var respJSON struct {
		Jobs  []jobInfo `json:"value"`
		Error string    `json:"error"`
	}
	if err := decodeJSON(resp.Body, &respJSON); err != nil {
		return nil, err
	}

	// handle status filter
	status := filters.String("status")
	if status != "" {
		respJSON.Jobs = jobFilter(respJSON.Jobs).Filter("status", status)
	}

	sort.Sort(jobSorter(respJSON.Jobs))
	if limit > 0 && limit < len(respJSON.Jobs) {
		respJSON.Jobs = respJSON.Jobs[:limit]
	}
	return respJSON.Jobs, err
}

func (p clientImpl) listBuilds(filters M) ([]jobInfo, error) {
	return p.listJobs(JobTypeBuild, filters)
}

func (p clientImpl) listDeployments(filters M) ([]jobInfo, error) {
	return p.listJobs(JobTypeDeployment, filters)
}

func (p clientImpl) listTests(filters M) ([]jobInfo, error) {
	return p.listJobs(JobTypeSimulation, filters)
}

func (p clientImpl) listGraphs(filters M) ([]jobInfo, error) {
	return p.listJobs(JobTypeGraph, filters)
}

func (p clientImpl) stopJob(eventType string, id string) error {
	var endpoint string
	switch eventType {
	case JobTypeSimulation:
		endpoint = endpoints.simulations.Events()
	case JobTypeDeployment:
		endpoint = endpoints.deployments.Events()
	default:
		endpoint = endpoints.builds.Events()
	}
	req := p.apiRequest(endpoint)
	req.param("id", id)
	reqBody := M{"status": "TERMINATING"}
	resp, err := req.Do("POST", reqBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errUnknownError
	}
	return nil
}

func inSlice(slice []string, val string) bool {
	for _, v := range slice {
		if val == v {
			return true
		}
	}
	return false
}
