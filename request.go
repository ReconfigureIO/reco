package reco

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Endpoint is reco client api endpoint.
type Endpoint string

// Append appends strings to endpoint path.
func (p Endpoint) Append(s ...string) string {
	return path.Join(string(p), path.Join(s...))
}

// String returns underlying string.
func (p Endpoint) String() string { return string(p) }

// Item returns item endpoint.
func (p Endpoint) Item() string {
	return p.Append("{id}")
}

// Input returns input endpoint.
func (p Endpoint) Input() string {
	return p.Append("{id}", "input")
}

// Log returns log endpoint.
func (p Endpoint) Log() string {
	return p.Append("{id}", "logs")
}

// Events returns events endpoint.
func (p Endpoint) Events() string {
	return p.Append("{id}", "events")
}

// Graph returns graph endpoint.
func (p Endpoint) Graph() string {
	return p.Append("{id}", "graph")
}

var (
	endpoints = struct {
		builds, deployments, projects, simulations, graphs Endpoint
	}{
		builds:      "/builds",
		simulations: "/simulations",
		projects:    "/projects",
		deployments: "/deployments",
		graphs:      "/graphs",
	}

	httpClient = &http.Client{
		Timeout: 0,
	}
)

// endpoint string, params urlParams, body io.Reader
type clientRequest struct {
	endpoint           string
	params             map[string]string
	username, password string
	jsonBody           bool
	queryParams        url.Values
	ctx                context.Context
}

func (p clientRequest) authenticated() bool {
	return p.username != "" && p.password != ""
}

func (p *clientRequest) param(key, value string) *clientRequest {
	if p.params == nil {
		p.params = make(map[string]string)
	}
	p.params[key] = value
	return p
}

func (p *clientRequest) queryParam(key, value string) *clientRequest {
	if p.queryParams == nil {
		p.queryParams = make(url.Values)
	}
	p.queryParams[key] = []string{value}
	return p
}

func (p *clientRequest) withContext(ctx context.Context) {
	p.ctx = ctx
}

// Do makes an http request with method and body. If body is not nil and not
// a io.Reader, body is encoded to JSON.
func (p *clientRequest) Do(method string, body interface{}) (*http.Response, error) {
	var reader io.Reader
	if !p.authenticated() {
		return nil, errAuthRequired
	}
	if _, ok := body.(io.Reader); !ok && body != nil {
		if r, err := jsonToReader(body); err == nil {
			reader = r
		}
	} else if ok {
		reader = body.(io.Reader)
	}
	for k, v := range p.params {
		p.endpoint = strings.Replace(p.endpoint, "{"+k+"}", v, -1)
	}
	endpoint := p.endpoint
	if len(p.queryParams) > 0 {
		endpoint = p.endpoint + "?" + p.queryParams.Encode()
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	if p.ctx != nil {
		req = req.WithContext(p.ctx)
	}
	if body != nil {
		if p.jsonBody {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/octect-stream")
		}
	}
	req.SetBasicAuth(p.username, p.password)
	resp, err := httpClient.Do(req)
	if resp != nil {
		switch resp.StatusCode {
		case 401, 403:
			err = errAuthFailed
		}
	} else if err != nil {
		err = errNetworkError
	}
	return resp, err
}

type apiResponse struct {
	ID  string `json:"id"`
	Job struct {
		Events []event `json:"events"`
	} `json:"job,omitempty"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Build struct {
		ID string `json:"id"`
		// workaround for deployments
		// TODO: fix on platform
		Project struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
	} `json:"build,omitempty"`
	// workaround for deployments
	// TODO: fix on platform
	Events  []event `json:"events,omitempty"`
	Command string  `json:"command,omitempty"`
}

type event struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Code      int       `json:"code"`
}
