# Finesse API

Utilities for work with the Unified Cisco Contact Center agents through Finesse API.

API allows actions per one agent or group of agents.

- get agent status
- login agent
- set agent ready state
- set agent not-ready state
- logout agent

## Connection
Program used connection to Finesse API and XMPP for notification.  
Utilizes ports:

- 8445 - HTTPS Finesse API
- 7443 - WSS Finesse XMPP over HTTP notification (for secure)
- 5222 - XMPP notification (non-secure - notice below)

***Notice:**
Cisco Finesse, Release 12.5(1) onward, the 5222 port (non-secure connection) is disabled
by default. Set the `utils finesse set_property webservices enableInsecureOpenfirePort` to true
to enable this port. For more information, see Service Properties section in
[Cisco Finesse Administration Guide](https://www.cisco.com/c/en/us/support/customer-collaboration/finesse/products-maintenance-guides-list.html).*

### Certificate
Program need to add the Finesse Notification certificate to their respective trust stores.

**Windows systems**:  
If you can use secure XMPP you must add valid server certificate to **Trusted Root Certification Authorities**.
- right-click on the DER file and select **Install certificate**
- select **Current User**
- select **Please all certificates in following store**
- click on **Browse...**
- select **Trusted Root Certification Authorities**
- finish

**Ubuntu**:
If you can use secure XMPP you must add valid server certificate to **Trusted Root Certification Authorities**.

```shell
sudo apt install ca-certificates
sudo cp finesse.pem /usr/local/share/ca-certificates/finesse.crt
sudo update-ca-certificates
```

#### How to download the certificate:

1. Sign in to the Cisco Unified Operating System Administration through the URL (https://FQDN:8443/cmplatform, where FQDN is the fully qualified domain name of the primary Finesse server and 8443 is the port number).
2. Click Security > Certificate Management.
3. Click Find to get the list of all the certificates.
4. In the Certificate List screen, choose Certificate from the Find Certificate List where drop-down menu, enter tomcat in the begins with option and click Find.
5. Click the FQDN link which appears in the Common Name column parallel to the listed tomcat certificate.
6. In the pop-up that appears, click the option Download .PEM or .DER File to save the file on your desktop.
 

System support security XMPP over HTTP (WSS).  
The current version does not support XMPP secure communication with the Finesse server.
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

Program use logger library "github.com/sirupsen/logrus".
