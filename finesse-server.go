package finesse_api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"time"
)

// SleepDurationMilliSecond Define delay between for one agent in milliseconds
// Default value is 1500 ms
var SleepDurationMilliSecond = 1500 * time.Millisecond

// FinesseServer Structure for finesse server data
type FinesseServer struct {
	id      string
	name    string
	port    int
	timeOut int
	agents  map[string]*FinesseAgent
	client  *http.Client
}

// NewFinesseServer Creating new Finesse server structure
func NewFinesseServer(server string, port int, time ...int) *FinesseServer {
	id := randomString()
	timeOut := 30
	if len(time) > 0 {
		timeOut = time[0]
	}
	return &FinesseServer{
		id:      id,
		name:    server,
		agents:  make(map[string]*FinesseAgent),
		port:    port,
		timeOut: timeOut,
		client:  nil,
	}
}

// AddAgent Adding new agent as part of finesse server
func (f *FinesseServer) AddAgent(agent *FinesseAgent) {
	f.agents[agent.LoginName] = agent
}

// LoginAgent Logging in agent with the agentName.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) LoginAgent(agentName string) bool {
	request := newFinesseRequest(f)
	agent, err := f.getAgentFromName(agentName)
	if err != nil {
		log.WithFields(log.Fields{logProc: "loginAgent", logId: request.id, logAgent: agentName}).
			Error(err)
		return false
	}
	log.WithFields(log.Fields{logProc: "loginAgent", logId: request.id, logAgent: agentName}).Tracef("success get agentId from name")
	if !f.getAgentId(agent) {
		return false
	}

	requestBody, err := agent.loginRequest().getUserRequest()
	if err != nil {
		log.WithFields(log.Fields{logProc: "loginAgent", logId: request.id, logAgent: agent.LoginName}).
			Errorf("not login agent %s on line %s. Prolem is %s", agent.LoginName, agent.Line, err)
		return false
	}
	response := request.doRequest("PUT", f.urlString(request.id, "User", agent.LoginId), agent, requestBody)
	defer response.close()
	msg, err := response.responseError()
	if err != nil {
		log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Error(msg)
		return false
	}
	log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).
		Tracef("agnet [%s] success login request", agent.LoginName)
	time.Sleep(SleepDurationMilliSecond)
	detail, err := f.stateAfterChange(agent.LoginName)
	if err != nil {
		return false
	}
	if !detail.IsLogIn() {
		log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).
			Errorf("agent %s not logged into server %s", agent.LoginName, f.name)
		return false
	}
	log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Debugf("agnet [%s] login success", agent.LoginName)
	return true
}

// LogoutAgent Logging out agent with the agentName.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) LogoutAgent(agentName string) bool {
	return f.doStateChange(agentName, "LOGOUT")
}

// ReadyAgent Setting agent with the agentName to ready state.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) ReadyAgent(agentName string) bool {
	return f.doStateChange(agentName, "READY")
}

// NotReadyAgent Setting agent with the agentName to not-ready state.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) NotReadyAgent(agentName string) bool {
	return f.doStateChange(agentName, "NOT_READY")
}

// NotReadyWithReasonAgent Setting agent with the agentName to not-ready state with reason.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) NotReadyWithReasonAgent(agentName string, reason int) bool {
	return f.doStateChange(agentName, "NOT_READY", reason)
}

// GetAgentStatusDetail Getting information about agent with the agentName.
// Agent name must be registered on this Finesse server with AddAgent function
func (f *FinesseServer) GetAgentStatusDetail(agentName string) (*UserDetailResponse, error) {
	return f.stateAfterChange(agentName)
}

func (f *FinesseServer) getAgentId(agent *FinesseAgent) bool {
	if len(agent.LoginId) <= 0 {
		request := newFinesseRequest(f)
		response := request.doRequest("GET", f.urlString(request.id, "User", agent.LoginName), agent, nil)
		defer response.close()
		time.Sleep(SleepDurationMilliSecond)
		msg, err := response.responseError()
		if err != nil {
			log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Error(msg)
			return false
		}
		log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Trace("success get data for agentId from name")
		data, err := newUserDetailResponse(response.GetResponseBody())
		if err != nil {
			log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Error(err)
			return false
		}
		if len(data.LoginId) <= 0 {
			log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Errorf("problem collect agentId from request")
			return false
		}
		agent.setAgentId(data.LoginId)
		log.WithFields(log.Fields{logProc: "loginAgent", logId: response.id, logAgent: agent.LoginName}).Tracef("success get agentId from name %s", data.LoginId)
	}
	return true
}

func (f *FinesseServer) stateAfterChange(agentName string) (*UserDetailResponse, error) {
	request := newFinesseRequest(f)
	agent, err := f.getAgentFromName(agentName)
	if err != nil {
		log.WithFields(log.Fields{logProc: "GetAgentStatusDetail", logId: request.id, logAgent: agentName}).
			Error(err)
		return nil, err
	}
	if !f.getAgentId(agent) {
		return nil, err
	}

	response := request.doRequest("GET", f.urlString(request.id, "User", agent.LoginId), agent, nil)
	msg, err := response.responseError()
	if err != nil {
		response.close()
		log.WithFields(log.Fields{logProc: "getAgentDetail", logId: response.id, logAgent: agent.LoginName}).Error(msg)
		return nil, err
	}
	data, err := newUserDetailResponse(response.GetResponseBody())
	if err != nil {
		log.WithFields(log.Fields{logProc: "getAgentDetail", logId: response.id, logAgent: agent.LoginName}).Error(err)
		return nil, err
	}
	return data, nil
}

func (f *FinesseServer) getAgentFromName(agentName string) (*FinesseAgent, error) {
	agent, found := f.agents[agentName]
	if !found {
		return nil, fmt.Errorf("agent [%s] not defined for server [%s]", agentName, f.name)
	}
	return agent, nil
}

func (f *FinesseServer) GetAgentsList() map[string]*FinesseAgent {
	return f.agents
}

func (f *FinesseServer) doStateChange(agentName string, requestState string, reason ...int) bool {
	request := newFinesseRequest(f)
	agent, err := f.getAgentFromName(agentName)
	if err != nil {
		log.WithFields(log.Fields{logProc: "doStateChange", logId: request.id, logAgent: agentName}).
			Error(err)
		return false
	}
	if !f.getAgentId(agent) {
		return false
	}

	var requestBody []byte
	if requestState == "NOT_READY" && len(reason) > 0 {
		state := userNotReadyWithReasonRequest{
			State:        requestState,
			ReasonCodeId: reason[0],
		}
		requestBody, err = state.getUserRequest()
	} else {
		state := userStateRequest{
			State: requestState,
		}
		requestBody, err = state.getUserRequest()
	}

	if err != nil {
		log.WithFields(log.Fields{logProc: "doStateChange", logId: request.id, logAgent: agent.LoginName, logNewState: requestState}).
			Errorf("change state to %s agent %s on line %s. Prolem is %s", requestState, agent.LoginName, agent.Line, err)
		return false
	}
	response := request.doRequest("PUT", f.urlString(request.id, "User", agent.LoginId), agent, requestBody)
	msg, err := response.responseError()
	if err != nil {
		response.close()
		log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: agent.LoginName, logNewState: requestState}).Error(msg)
		return false
	}
	log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: agent.LoginName}).
		Tracef("agnet [%s] state change request", agent.LoginName)
	time.Sleep(SleepDurationMilliSecond)
	detail, err := f.stateAfterChange(agentName)
	if err != nil {
		return false
	}
	if detail.State != requestState {
		log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: agent.LoginName}).
			Errorf("agent %s not change state to %s into server %s", agent.LoginName, requestState, f.name)
		return false
	}
	log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: agent.LoginName}).Debugf("agnet [%s] state change success", agent.LoginName)
	return true
}

func (f *FinesseServer) setHttpClient(client *http.Client) {
	f.client = client
}

func (f *FinesseServer) urlString(rId string, pathPart ...string) string {
	restPath := "/finesse/api"
	if len(pathPart) > 0 {
		for _, s := range pathPart {
			restPath = path.Join(restPath, s)
		}
	}
	var url string
	if f.port != 80 {
		url = fmt.Sprintf("https://%s:%d%s", f.name, f.port, restPath)
	} else {
		url = fmt.Sprintf("https://%s%s", f.name, restPath)
	}
	log.WithFields(log.Fields{logProc: "urlString", logServer: f.name, logId: rId}).Tracef("Request URI: %s", url)
	return url
}
