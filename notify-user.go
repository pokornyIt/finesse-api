package finesse_api

import "encoding/xml"

type XmppUser struct {
	//XMLName      xml.Name `xml:"User"`
	Dialogs      string `xml:"dialogs"`
	Extension    string `xml:"extension"`
	FirstName    string `xml:"firstName"`
	LastName     string `xml:"lastName"`
	LoginId      string `xml:"loginId"`
	LoginName    string `xml:"loginName"`
	MediaType    string `xml:"mediaType"`
	ReasonCodeId string `xml:"reasonCodeId"`
	ReasonCode   struct {
		Category string `xml:"category"`
		URL      string `xml:"uri"`
		Code     string `xml:"code"`
		Label    string `xml:"label"`
		ForAll   bool   `xml:"forAll"`
		Id       int    `xml:"id"`
	} `xml:"ReasonCode"`
	Roles struct {
		Role []string `xml:"role"`
	} `xml:"roles"`
	Settings struct {
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
		Team []struct {
			Id   int    `xml:"id"`
			Name string `xml:"name"`
			URI  string `xml:"uri"`
		} `xml:"Team"`
	} `xml:"teams"`
	MobileAgent struct {
		Mode       string `xml:"mode"`
		DialNumber string `xml:"dialNumber"`
	} `xml:"mobileAgent"`
	ActiveDeviceId string `xml:"activeDeviceId"`
	Devices        struct {
		Device []struct {
			DeviceId       string `xml:"deviceId"`
			DeviceType     string `xml:"deviceType"`
			DeviceTypeName string `xml:"deviceTypeName"`
		} `xml:"device"`
	} `xml:"devices"`
}

func newXmppUser(data string) (*XmppUser, error) {
	var u XmppUser
	buffer := []byte(data)
	err := xml.Unmarshal(buffer, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
