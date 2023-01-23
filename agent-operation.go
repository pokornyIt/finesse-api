package finesse_api

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	XmppTimeout                 = 20
	TypeErrorNoError            = 0
	TypeErrorWrongState         = 1
	TypeErrorRequest            = 2
	TypeErrorResponse           = 3
	TypeErrorNotifyTimeout      = 4
	TypeErrorAnalyzeResponse    = 5
	TypeErrorUnknownBulkCommand = 6
	TypeErrorNoStatus           = 7
)

type OperationError struct {
	Type  int
	Error error
}

func (a *Agent) Login() OperationError {
	if a.lastStatus.State != AgentStateLogout {
		return OperationError{
			Type:  TypeErrorWrongState,
			Error: fmt.Errorf("agent [%s] is in [%s] state", a.LoginName, a.lastStatus.State),
		}
	}
	return a.doStateChange(AgentStateLogin)
}

func (a *Agent) Logout(forceLogout ...bool) OperationError {
	force := false
	if len(forceLogout) > 0 {
		force = forceLogout[0]
	}
	if a.lastStatus.State == AgentStateReady && force {
		errOp := a.NotReady()
		if errOp.Type != TypeErrorNoError {
			return errOp
		}
	}
	if a.lastStatus.State != AgentStateNotReady {
		return OperationError{
			Type:  TypeErrorWrongState,
			Error: fmt.Errorf("agent [%s] is in [%s] state and not possible logout", a.LoginName, a.lastStatus.State),
		}
	}
	return a.doStateChange(AgentStateLogout)
}

func (a *Agent) Ready(forceReady ...bool) OperationError {
	force := false
	if len(forceReady) > 0 {
		force = forceReady[0]
	}
	_, ok := AgentReadyStates[a.lastStatus.State]
	if ok {
		return OperationError{
			Type:  TypeErrorWrongState,
			Error: fmt.Errorf("agent [%s] is in [%s] state and not possible switch to ready", a.LoginName, a.lastStatus.State),
		}
	}

	if a.lastStatus.State == AgentStateLogout && force {
		errOp := a.Login()
		if errOp.Type != TypeErrorNoError {
			return errOp
		}
	}

	return a.doStateChange(AgentStateReady)
}

func (a *Agent) NotReady() OperationError {
	_, ok := AgentReadyStates[a.lastStatus.State]
	if !ok {
		return OperationError{
			Type:  TypeErrorWrongState,
			Error: fmt.Errorf("agent [%s] is in [%s] state and not possible switch to ready", a.LoginName, a.lastStatus.State),
		}
	}
	return a.doStateChange(AgentStateNotReady)
}

func (a *Agent) doStateChange(requestState string, reason ...int) OperationError {
	var err error
	request := a.newAgentRequest()
	var requestBody []byte
	if requestState == AgentStateNotReady && len(reason) > 0 {
		state := userStateWithReasonRequest{
			State:        requestState,
			ReasonCodeId: reason[0],
		}
		requestBody, err = state.getUserRequest()
	} else if requestState == AgentStateLogout && len(reason) > 0 {
		state := userStateWithReasonRequest{
			State:        requestState,
			ReasonCodeId: reason[0],
		}
		requestBody, err = state.getUserRequest()

	} else if requestState == AgentStateLogin {
		state := userLoginRequest{
			XMLName:   xml.Name{},
			Text:      "",
			State:     AgentStateLogin,
			Extension: a.Line,
		}
		requestBody, err = state.getUserRequest()
	} else {
		state := userStateRequest{
			State: requestState,
		}
		requestBody, err = state.getUserRequest()
	}

	if err != nil {
		log.WithFields(log.Fields{logProc: "doStateChange", logId: request.id, logAgent: a.LoginName, logNewState: requestState}).
			Errorf("change state to %s agent %s on line %s. Prolem is %s", requestState, a.LoginName, a.Line, err)
		return OperationError{
			Type:  TypeErrorRequest,
			Error: err,
		}
	}
	// clean queue https://stackoverflow.com/a/26143288/4074126
	for len(a.response) > 0 {
		data := <-a.response
		log.WithFields(log.Fields{logProc: "doStateChange", logId: request.id, logAgent: a.LoginName, logNewState: requestState}).Debugf("remove data from channel [%s]", data)
	}

	response := request.doRequest("PUT", a.server.urlString(request.id, "User", a.LoginId), requestBody)
	msg, err := response.responseError()
	if err != nil {
		response.close()
		log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: a.LoginName, logNewState: requestState}).Error(msg)
		return OperationError{
			Type:  TypeErrorResponse,
			Error: err,
		}
	}
	log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: a.LoginName, logNewState: requestState}).
		Tracef("agnet [%s] state change request", a.LoginName)

	select {
	case status := <-a.response:
		log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: a.LoginName, logNewState: requestState}).Tracef("get message from XMPP notification")
		s, e := a.analyzeResponse(status)
		if e != nil {
			log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: a.LoginName, logNewState: requestState}).Error(e)
			return OperationError{
				Type:  TypeErrorAnalyzeResponse,
				Error: e,
			}
		} else {
			a.lastStatus = s
		}
		return OperationError{
			Type:  TypeErrorNoError,
			Error: nil,
		}
	case <-time.After(time.Duration(XmppTimeout) * time.Second):
		log.WithFields(log.Fields{logProc: "doStateChange", logId: response.id, logAgent: a.LoginName, logNewState: requestState}).Error("collect notify response form XMPP timeouts")
		return OperationError{
			Type:  TypeErrorNotifyTimeout,
			Error: fmt.Errorf("timeout collect notify response form XMPP for agnet [%s]", a.LoginName),
		}
	}
}
