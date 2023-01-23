package finesse_api

type XmppTeam struct {
	URI   string `xml:"uri"`
	ID    string `xml:"id"`
	Name  string `xml:"name"`
	Users struct {
		User []struct {
			URI             string `xml:"uri"`
			LoginId         string `xml:"loginId"`
			FirstName       string `xml:"firstName"`
			LastName        string `xml:"lastName"`
			Dialogs         string `xml:"dialogs"`
			Extension       string `xml:"extension"`
			PendingState    string `xml:"pendingState"`
			State           string `xml:"state"`
			StateChangeTime string `xml:"stateChangeTime"`
			ReasonCode      struct {
				Category string `xml:"category"`
				Code     string `xml:"code"`
				Label    string `xml:"label"`
				ID       string `xml:"id"`
				URI      string `xml:"uri"`
			} `xml:"reasonCode"`
		} `xml:"User"`
	} `xml:"users"`
}

type XmppTeamMessage struct {
	URI       string `xml:"uri"`
	ID        string `xml:"id"`
	CreatedBy struct {
		ID        string `xml:"id"`
		FirstName string `xml:"firstName"`
		LastName  string `xml:"lastName"`
	} `xml:"createdBy"`
	CreatedAt string `xml:"createdAt"`
	Duration  string `xml:"duration"`
	Content   string `xml:"content"`
	Teams     struct {
		Team []string `xml:"team"`
	} `xml:"teams"`
}
