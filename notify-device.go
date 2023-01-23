package finesse_api

type XmppDevices struct {
	Device []struct {
		DeviceId       string `xml:"deviceId"`
		DeviceType     string `xml:"deviceType"`
		DeviceTypeName string `xml:"deviceTypeName"`
	} `xml:"Device"`
}
