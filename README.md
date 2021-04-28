# Finesse API

Utilities for work with the Unified Cisco Contact Center agents through Finesse API.

API allows actions per one agent or group of agents.

- get agent status
- login agent
- set agent ready state
- set agent not-ready state
- logout agent

## Limitation
The current version does not use XMPP communication with the Finesse server.
Program tested on version Finesse 12.5.

## How use

For any operation is necessary to create a server structure with an address and port.
It is necessary to register the agents with which the operations will take place on the server.


```go
// import finesse_api library
import (
	api "github.com/pokornyIt/finesse-api"
)

// create Finesse server object
server := finesse_api.NewFinesseServer("finesse.server.fqdn", 8435)

// add agents to server
agent := api.NewAgentName("Name1", "Password", "1000")
server.AddAgent(agent)
agent := api.NewAgentName("Name2", "Password", "1001")
server.AddAgent(agent)
agent := api.NewAgentName("Name3", "Password", "1002")
server.AddAgent(agent)

// get status for all defined agent
states, err = server.GetStateAgentsParallel()

// login agent Name2 to system
state := server.LoginAgent("Name2")

// login all agents and set it to ready state
states, err = server.ReadyAgentsParallelWithStatus(true)
```

Program use standard logger library "github.com/sirupsen/logrus".
