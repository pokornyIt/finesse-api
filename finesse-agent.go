package finesse_api

import (
	"encoding/xml"
)

// FinesseAgent Login information of agent
type FinesseAgent struct {
	LoginName string // login name
	LoginId   string // login ID
	Password  string // password
	Line      string // phone line
}

// NewAgent Creating agent login data
func NewAgent(id string, name string, pwd string, line string) *FinesseAgent {
	return &FinesseAgent{
		LoginId:   id,
		LoginName: name,
		Password:  pwd,
		Line:      line,
	}
}

// NewAgentName Creating agent login data without "agent Id" necessary for requests
func NewAgentName(name string, pwd string, line string) *FinesseAgent {
	return NewAgent("", name, pwd, line)
}

func (a *FinesseAgent) loginRequest() *userLoginRequest {
	return &userLoginRequest{
		XMLName:   xml.Name{},
		Text:      "",
		State:     AgentStateLogin,
		Extension: a.Line,
	}
}

func (a *FinesseAgent) setAgentId(newId string) {
	a.LoginId = newId
}
