package finesse_api

import (
	"encoding/xml"
)

// userLoginRequest structure for agent login
type userLoginRequest struct {
	XMLName   xml.Name `xml:"User"`
	Text      string   `xml:",chardata"`
	State     string   `xml:"state"`
	Extension string   `xml:"extension"`
}

// userStateRequest structure for agent logout
type userStateRequest struct {
	XMLName xml.Name `xml:"User"`
	Text    string   `xml:",chardata"`
	State   string   `xml:"state"`
}

// userNotReadyWithReasonRequest structure for agent logout
type userNotReadyWithReasonRequest struct {
	XMLName      xml.Name `xml:"User"`
	Text         string   `xml:",chardata"`
	State        string   `xml:"state"`
	ReasonCodeId int      `xml:"reasonCodeId"`
}

type userRequest interface {
	getUserRequest() ([]byte, error)
}

func (u *userLoginRequest) getUserRequest() ([]byte, error) {
	data, err := xml.Marshal(u)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *userStateRequest) getUserRequest() ([]byte, error) {
	data, err := xml.Marshal(u)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *userNotReadyWithReasonRequest) getUserRequest() ([]byte, error) {
	data, err := xml.Marshal(u)
	if err != nil {
		return nil, err
	}
	return data, nil
}

