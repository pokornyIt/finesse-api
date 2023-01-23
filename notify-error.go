package finesse_api

type XmppErrors struct {
	ApiErrors []XmppError `xml:"apiError"`
}

type XmppError struct {
	PeripheralErrorCode int    `xml:"peripheralErrorCode,omitempty"`
	ErrorType           string `xml:"errorType,omitempty"`
	ErrorMessage        string `xml:"errorMessage,omitempty"`
	PeripheralErrorText string `xml:"peripheralErrorText,omitempty"`
	PeripheralErrorMsg  string `xml:"peripheralErrorMsg,omitempty"`
	ErrorData           int    `xml:"errorData,omitempty"`
}
