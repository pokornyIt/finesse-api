package finesse_api

type XmppQueue struct {
	URI        string `xml:"uri"`
	Name       string `xml:"name"`
	Statistics struct {
		CallsInQueue                  string `xml:"callsInQueue"`
		StartTimeOfLongestCallInQueue string `xml:"startTimeOfLongestCallInQueue"`
		AgentsReady                   string `xml:"agentsReady"`
		AgentsNotReady                string `xml:"agentsNotReady"`
		AgentsBusyOther               string `xml:"agentsBusyOther"`
		AgentsLoggedOn                string `xml:"agentsLoggedOn"`
		AgentsTalkingInbound          string `xml:"agentsTalkingInbound"`
		AgentsTalkingOutbound         string `xml:"agentsTalkingOutbound"`
		AgentsTalkingInternal         string `xml:"agentsTalkingInternal"`
		AgentsWrapUpNotReady          string `xml:"agentsWrapUpNotReady"`
		AgentsWrapUpReady             string `xml:"agentsWrapUpReady"`
	} `xml:"statistics"`
}
