package finesse_api

type XmppDialogs struct {
	Dialogs []XmppDialog `xml:"Dialog"`
}

type XmppDialog struct {
	AssociatedDialogUri string `xml:"associatedDialogUri"`
	FromAddress         string `xml:"fromAddress"`
	ID                  string `xml:"id"`
	SecondaryId         string `xml:"secondaryId"`
	MediaProperties     struct {
		MediaId                string `xml:"mediaId"`
		DNIS                   string `xml:"DNIS"`
		CallType               string `xml:"callType"`
		DialedNumber           string `xml:"dialedNumber"`
		OutboundClassification string `xml:"outboundClassification"`
		CallVariables          struct {
			CallVariable []struct {
				Name  string `xml:"name"`
				Value string `xml:"value"`
			} `xml:"CallVariable"`
		} `xml:"callvariables"`
		QueueNumber        string `xml:"queueNumber"`
		QueueName          string `xml:"queueName"`
		CallKeyCallId      string `xml:"callKeyCallId"`
		CallKeySequenceNum string `xml:"callKeySequenceNum"`
		CallKeyPrefix      string `xml:"callKeyPrefix"`
	} `xml:"mediaProperties"`
	MediaType    string `xml:"mediaType"`
	Participants struct {
		Participant struct {
			Actions struct {
				Action []string `xml:"action"`
			} `xml:"actions"`
			MediaAddress     string `xml:"mediaAddress"`
			MediaAddressType string `xml:"mediaAddressType"`
			StartTime        string `xml:"startTime"`
			State            string `xml:"state"`
			StateCause       string `xml:"stateCause"`
			StateChangeTime  string `xml:"stateChangeTime"`
		} `xml:"Participant"`
	} `xml:"participants"`
	State     string `xml:"state"`
	ToAddress string `xml:"toAddress"`
	URI       string `xml:"uri"`
}
