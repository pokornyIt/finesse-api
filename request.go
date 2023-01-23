package finesse_api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// AgentRequest Structure for one API request
type AgentRequest struct {
	id        string
	loginName string
	password  string
	line      string
	client    *http.Client
	server    *Server
	request   *http.Request
}

// setHeader create request header
func (f *AgentRequest) setHeader() {
	if f.request.Method != "GET" {
		f.request.Header.Set("Content-Type", "application/xml")
	}
	f.request.Header.Set("User-Agent", "Finesse/1.0")
	f.request.Header.Set("Accept", "*/*")
	f.request.Header.Set("Cache-Control", "no-cache")
	//f.request.Header.Set("Pragma", "no-cache")
	f.request.Header.Set("RequestId", f.id)
	f.request.Host = f.server.name
	f.request.SetBasicAuth(f.loginName, f.password)
}

// httpClient prepare httpClient for request
func (f *AgentRequest) httpClient() {
	if f.client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		f.client = &http.Client{Timeout: time.Duration(f.server.timeOut) * time.Second, Transport: tr}
		log.WithFields(log.Fields{logProc: "httpClient", logId: f.id}).Debugf("prepare HTTP client for server [%s] in request", f.server.name)
	}
}

// doRequest process one request
func (f *AgentRequest) doRequest(method string, url string, data []byte) *AgentResponse {
	log.WithFields(log.Fields{logProc: "doRequest", logId: f.id, logRequestType: method, logBody: string(data)}).Tracef("start process request [%s %s]", method, url)
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		log.WithFields(log.Fields{logProc: "doRequest", logId: f.id}).Errorf(
			"problem create [%s %s] request for [%s] agent with error %s", method, url, f.loginName, err)
	}
	f.request = request
	f.setHeader()
	f.httpClient()
	resp, err := f.client.Do(f.request)
	if err != nil {
		r := fmt.Sprintf("problem request [%s %s]", f.request.Method, f.request.URL)
		log.WithFields(log.Fields{logProc: "doRequest", logId: f.id}).Error(r)
		return f.newResponse(resp, err, r)
	}
	return f.newResponse(resp, nil, "")
}

// newResponse Create new response structure
func (f *AgentRequest) newResponse(response *http.Response, e error, message string) *AgentResponse {
	r := new(AgentResponse)
	r.id = f.id
	r.response = response
	r.err = e
	r.lastMessage = message
	if response != nil {
		r.statusCode = response.StatusCode
		r.statusMessage = response.Status
	} else {
		r.statusCode = 500
		r.statusMessage = "500 Problem Connect to server"
	}
	log.WithFields(log.Fields{logProc: "NewResponse", logId: r.id}).Tracef("response with status [%s]", r.statusMessage)
	return r
}
