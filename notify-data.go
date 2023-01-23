package finesse_api

type XmppUpdate struct {
	Event     string `xml:"event"`
	RequestId string `xml:"requestId"`
	Source    string `xml:"source"`
	Data      struct {
		User        XmppUser        `xml:"user,omitempty"`
		Error       XmppErrors      `xml:"apiErrors,omitempty"`
		Dialogs     XmppDialogs     `xml:"Dialog,omitempty"`
		Devices     XmppDevices     `xml:"Devices,omitempty"`
		Queue       XmppQueue       `xml:"Queue,omitempty"`
		Team        XmppTeam        `xml:"Team,omitempty"`
		TeamMessage XmppTeamMessage `xml:"TeamMessage,omitempty"`
	} `xml:"data"`
}
