package finesse_api

import (
	"encoding/xml"
	"fmt"
)

// UserDetailResponse Structure holds agent status
// Contains all possible values returned in the get agent state
type UserDetailResponse struct {
	XMLName      xml.Name `xml:"User"`
	Text         string   `xml:",chardata"`
	Dialogs      string   `xml:"dialogs"`
	Extension    string   `xml:"extension"`
	FirstName    string   `xml:"firstName"`
	LastName     string   `xml:"lastName"`
	LoginId      string   `xml:"loginId"`
	LoginName    string   `xml:"loginName"`
	MediaType    string   `xml:"mediaType"`
	ReasonCodeId string   `xml:"reasonCodeId"`
	ReasonCode   struct {
		Text     string `xml:",chardata"`
		Category string `xml:"category"`
		URL      string `xml:"uri"`
		Code     string `xml:"code"`
		Label    string `xml:"label"`
		ForAll   bool   `xml:"forAll"`
		Id       int    `xml:"id"`
	} `xml:"ReasonCode"`
	Roles struct {
		Text string   `xml:",chardata"`
		Role []string `xml:"role"`
	} `xml:"roles"`
	Settings struct {
		Text             string `xml:",chardata"`
		WrapUpOnIncoming string `xml:"wrapUpOnIncoming"`
		WrapUpOnOutgoing string `xml:"wrapUpOnOutgoing"`
		DeviceSelection  string `xml:"deviceSelection"`
	} `xml:"settings"`
	State           string `xml:"state"`
	StateChangeTime string `xml:"stateChangeTime"`
	PendingState    string `xml:"pendingState"`
	TeamId          string `xml:"teamId"`
	TeamName        string `xml:"teamName"`
	SkillTargetId   string `xml:"skillTargetId"`
	URI             string `xml:"uri"`
	Teams           struct {
		Text string `xml:",chardata"`
		Team []struct {
			Text string `xml:",chardata"`
			Id   int    `xml:"id"`
			Name string `xml:"name"`
			URI  string `xml:"uri"`
		} `xml:"Team"`
	} `xml:"teams"`
	MobileAgent struct {
		Text       string `xml:",chardata"`
		Mode       string `xml:"mode"`
		DialNumber string `xml:"dialNumber"`
	} `xml:"mobileAgent"`
	ActiveDeviceId string `xml:"activeDeviceId"`
	Devices        struct {
		Text   string `xml:",chardata"`
		Device []struct {
			Text           string `xml:",chardata"`
			DeviceId       string `xml:"deviceId"`
			DeviceType     string `xml:"deviceType"`
			DeviceTypeName string `xml:"deviceTypeName"`
		} `xml:"device"`
	} `xml:"devices"`
}

func newUserDetailResponse(data string) (*UserDetailResponse, error) {
	var u UserDetailResponse
	buffer := []byte(data)
	err := xml.Unmarshal(buffer, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ToString Returns the basic data from the status response as a printable string
func (u *UserDetailResponse) ToString() string {
	s := fmt.Sprintf("Dialogs: %s\r\n", u.Dialogs)
	s = fmt.Sprintf("%sExtension: %s\r\n", s, u.Extension)
	s = fmt.Sprintf("%sFirst Name: %s\r\n", s, u.FirstName)
	s = fmt.Sprintf("%sLast Name: %s\r\n", s, u.LastName)
	s = fmt.Sprintf("%sLogin ID: %s\r\n", s, u.LoginId)
	s = fmt.Sprintf("%sLogin name: %s\r\n", s, u.LoginName)
	s = fmt.Sprintf("%sState: %s\r\n", s, u.State)
	s = fmt.Sprintf("%sTeam name: %s\r\n", s, u.TeamName)
	s = fmt.Sprintf("%sTeam ID: %s\r\n", s, u.TeamId)
	s = fmt.Sprintf("%sPending state: %s\r\n", s, u.PendingState)
	s = fmt.Sprintf("%sReason code ID: %s\r\n", s, u.ReasonCodeId)
	s = fmt.Sprintf("%sRole: %s\n\r", s, u.getRoles())
	s = fmt.Sprintf("%sTeams: %s\n\r", s, u.getTeams())

	return s
}

// ToStingSimple Returns agent name with current state
func (u *UserDetailResponse) ToStingSimple() string {
	return fmt.Sprintf("Agent: %-30s State: %-15s Pending state %s", u.LoginName, u.State, u.PendingState)
}

func (u *UserDetailResponse) getRoles() string {
	a := ""
	sep := ""
	for _, role := range u.Roles.Role {
		a = fmt.Sprintf("%s%s%s", a, sep, role)
		sep = ", "
	}
	return a
}

func (u *UserDetailResponse) getTeams() string {
	a := ""
	sep := ""
	for _, team := range u.Teams.Team {
		a = fmt.Sprintf("%s%s%s", a, sep, team.Name)
		sep = ", "
	}
	return a
}

// IsLogIn Is agent logged in
func (u *UserDetailResponse) IsLogIn() bool {
	return u.State != AgentStateLogout
}

// IsPossibleToLogout Is agent ready for logout
func (u *UserDetailResponse) IsPossibleToLogout() bool {
	return u.State == AgentStateNotReady
}
