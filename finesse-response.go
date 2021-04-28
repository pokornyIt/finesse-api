package finesse_api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// FinesseResponse Structure for one API response
type FinesseResponse struct {
	id            string
	response      *http.Response
	err           error
	lastMessage   string
	body          string
	statusCode    int
	statusMessage string
}

// NewFinesseResponse Create new response structure
func (f *FinesseRequest) NewFinesseResponse(response *http.Response, e error, message string) *FinesseResponse {
	r := new(FinesseResponse)
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
	log.WithFields(log.Fields{logProc: "NewFinesseResponse", logId: r.id}).Tracef("response with status [%s]", r.statusMessage)
	return r
}

func (f *FinesseResponse) close() {
	if f.response != nil {
		if f.response.Body != nil {
			_ = f.response.Body.Close()
		}
		f.response = nil
	}
}

func (f *FinesseResponse) responseError() (string, error) {
	if f.statusCode >= 200 && f.statusCode <= 299 {
		return "", nil
	}
	if len(f.lastMessage) > 0 {
		return f.lastMessage, fmt.Errorf(f.lastMessage)
	}
	return f.lastMessage, fmt.Errorf("reponse with error [%s]", f.statusMessage)
}

// GetResponseBody Read API response body
func (f *FinesseResponse) GetResponseBody() string {
	if f.response == nil {
		return f.body
	}
	err := f.responseReturnData()
	if err != nil {
		f.err = err
	}
	return f.body
}

func (f *FinesseResponse) responseReturnData() error {
	log.WithFields(log.Fields{logProc: "responseReturnData", logId: f.id, logHttpStatus: f.response.Status}).
		Debugf("response status is [%s]", f.response.Status)
	bodies, err := io.ReadAll(f.response.Body)
	_ = f.response.Body.Close()
	f.body = ""

	if err != nil {
		log.WithFields(log.Fields{logProc: "responseReturnData", logId: f.id}).Errorf("problem get body from response [%s]", err)
		return err
	}
	f.body = string(bodies)
	log.WithFields(log.Fields{logProc: "responseReturnData", logId: f.id}).Tracef("body read success [%s %s]", f.response.Request.Method, f.response.Request.URL)
	if f.statusCode > 299 {
		return fmt.Errorf(f.statusMessage)
	}
	return nil
}
