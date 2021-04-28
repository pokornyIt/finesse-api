package finesse_api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// FinesseRequest Structure for one API request
type FinesseRequest struct {
	id      string
	client  *http.Client
	server  *FinesseServer
	request *http.Request
}

func newFinesseRequest(server *FinesseServer) *FinesseRequest {
	r := FinesseRequest{
		id:      randomString(),
		client:  server.client,
		server:  server,
		request: nil,
	}
	log.WithFields(log.Fields{logProc: "NewRequest", logId: r.id, logServer: r.server.name}).Tracef("prepare new request for server [%s]", server.name)
	return &r
}

func (f *FinesseRequest) setHeader(agent *FinesseAgent) {
	if f.request.Method != "GET" {
		f.request.Header.Set("Content-Type", "application/xml")
	}
	f.request.Header.Set("User-Agent", "Finesse/1.0")
	f.request.Header.Set("Accept", "*/*")
	f.request.Header.Set("Cache-Control", "no-cache")
	//f.request.Header.Set("Pragma", "no-cache")
	f.request.Header.Set("RequestId", f.id)
	f.request.Host = f.server.name
	f.request.SetBasicAuth(agent.LoginName, agent.Password)
}

func (f *FinesseRequest) httpClient() {
	if f.client == nil {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		f.client = &http.Client{Timeout: time.Duration(f.server.timeOut) * time.Second, Transport: tr}
		log.WithFields(log.Fields{logProc: "httpClient", logId: f.id}).Debugf("prepare HTTP client for server [%s] in request", f.server.name)
	}
}

func (f *FinesseRequest) doRequest(method string, url string, agent *FinesseAgent, data []byte) *FinesseResponse {
	log.WithFields(log.Fields{logProc: "doRequest", logId: f.id, logRequestType: method, logBody: string(data)}).Tracef("start process request [%s %s]", method, url)
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		log.WithFields(log.Fields{logProc: "doRequest", logId: f.id}).Errorf(
			"problem create [%s %s] request for [%s] agent with error %s", method, url, agent.LoginName, err)
	}
	f.request = request
	f.setHeader(agent)
	f.httpClient()
	resp, err := f.client.Do(f.request)
	if err != nil {
		r := fmt.Sprintf("problem request [%s %s]", f.request.Method, f.request.URL)
		log.WithFields(log.Fields{logProc: "doRequest", logId: f.id}).Error(r)
		return f.NewFinesseResponse(resp, err, r)
	}
	return f.NewFinesseResponse(resp, nil, "")
}
