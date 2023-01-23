package finesse_api

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const (
	XmppMessageBuffer = 10 // define channel buffer for collect message from Xmpp notify service
)

type Agent struct {
	LoginName            string              // login name
	LoginId              string              // login ID
	Password             string              // password
	Line                 string              // phone line
	lastStatus           *XmppUser           // latest agent response
	httpClient           *http.Client        // prepared HTTP client
	streamManagerService *xmpp.StreamManager // XMPP scream manager for user
	ctx                  context.Context     // context for graceful shutdown of notify subroutine
	server               *Server             // associate finesse server
	response             chan string         // channel for get strings
}

// NewAgentNotify create new agent object, but not create/start any additional service
//
// Better way is use function Server.CreateAgent, this creates agent and start necessary function
func NewAgentNotify(ctx context.Context, name string, pwd string, line string, server *Server) *Agent {
	return &Agent{
		LoginName:            name,
		LoginId:              "",
		Password:             pwd,
		Line:                 line,
		lastStatus:           nil,
		httpClient:           nil,
		streamManagerService: nil,
		server:               server,
		ctx:                  ctx,
		response:             make(chan string, XmppMessageBuffer),
	}
}

func (a *Agent) newAgentRequest() *AgentRequest {
	r := AgentRequest{
		id:        randomString(),
		client:    a.server.getHttpClient(),
		server:    a.server,
		request:   nil,
		loginName: a.LoginName,
		password:  a.Password,
		line:      a.Line,
	}
	log.WithFields(log.Fields{logProc: "NewRequest", logId: r.id, logServer: r.server.name}).Tracef("prepare new request for server [%s]", a.server.name)
	return &r
}

// getId read ID from Finesse and store it if OK
func (a *Agent) getId() error {
	if a.LoginId != "" {
		return nil
	}
	request := a.newAgentRequest()
	response := request.doRequest("GET", a.server.urlString(request.id, "User", a.LoginName), nil)
	defer response.close()
	msg, err := response.responseError()
	if err != nil {
		log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Error(msg)
		return err
	}
	log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Trace("success get data for agentId from name")
	data, err := newXmppUser(response.GetResponseBody())
	if err != nil {
		log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Error(err)
		return err
	}
	if len(data.LoginId) <= 0 {
		log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Errorf("problem collect agentId from request")
		return fmt.Errorf("agent ID is empty for agent name %s", a.LoginName)
	}
	a.lastStatus = data
	a.LoginId = data.LoginId
	log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Tracef("collect agentId [%s] for agent [%s]", data.LoginId, a.LoginName)
	return nil
}

func (a *Agent) getDomain() string {
	if strings.Contains(a.LoginName, "@") {
		sp := strings.Split(a.LoginName, "@")
		if len(sp) == 2 {
			return sp[1]
		}
	}
	return a.server.getDomain()
}

func (a *Agent) StartXmpp() error {
	if a.streamManagerService != nil {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("start finesse_notifier - XMPP notifier is ready ")
		return nil
	}
	if a.LoginId == "" {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Errorf("XMPP not start missing agent login ID")
		return fmt.Errorf("XMPP not start missing agent login ID")
	}

	log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("start finesse_notifier")

	// setup WSS or XMPP connection parameters
	t := &tls.Config{InsecureSkipVerify: a.server.ignore}
	domain := a.getDomain()
	server := fmt.Sprintf("wss://%s:%d/ws/", a.server.name, a.server.xmppPort)

	if a.server.insecureXmpp {
		server = fmt.Sprintf("%s:%d", a.server.name, a.server.xmppPort)
	}
	log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).
		Debugf("finesse_notifier server [%s] with domain [%s] ignore certificate problem [%t]", server, domain, a.server.ignore)

	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address:   server,
			Domain:    domain,
			TLSConfig: t,
		},
		Jid:        fmt.Sprintf("%s@%s", a.LoginId, a.server.name),
		Credential: xmpp.Password(a.Password),
		//StreamLogger: os.Stdout,
		Insecure: a.server.ignore,
	}

	//goland:noinspection HttpUrlsUsage
	stanza.TypeRegistry.MapExtension(stanza.PKTMessage, xml.Name{Space: "http://jabber.org/protocol/pubsub#event", Local: "event"}, stanza.PubSubEvent{})
	router := xmpp.NewRouter()
	//router.HandleFunc("message", a.messageHandler)
	router.HandleFunc("message", func(s xmpp.Sender, p stanza.Packet) {
		log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Trace("handle XMPP message stream")
		msg, ok := p.(stanza.Message)
		if !ok {
			log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Tracef("ignore packet %T", p)
			return
		}
		if len(msg.Extensions) > 0 {
			for _, extension := range msg.Extensions {
				if "*stanza.PubSubEvent" == reflect.TypeOf(extension).String() {
					ext := extension.(*stanza.PubSubEvent)
					if "*stanza.ItemsEvent" == reflect.TypeOf(ext.EventElement).String() {
						element := ext.EventElement.(*stanza.ItemsEvent)
						for _, item := range element.Items {
							log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Trace("success accept message")

							select {
							case a.response <- item.Any.Content:
								log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Trace("success send new data into buffered queue")
							case <-time.After(time.Duration(20) * time.Second):
								log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Warnf("timeout send message to queue")
							default:
								log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Warnf("response buffered queue is full. Data lost!")
							}
						}
					} else {
						log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).
							Warnf("PubSubEvent doesnt contains unexpected type [%s]", reflect.TypeOf(ext.EventElement).String())
					}
				} else {
					log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).
						Warnf("unknown XMPP extension type [%s]", reflect.TypeOf(extension).String())
				}
			}
		} else {
			log.WithFields(log.Fields{logProc: "messageHandler", logAgent: a.LoginName}).Warnf("XMPP message without extension type")
		}
	})

	client, err := xmpp.NewClient(&config, router, func(err error) {
		//log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Errorf("client error handler messages - %s", err)
	})
	log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Debugf("prepare XMPP client for agent [%s]", a.LoginName)

	a.streamManagerService = xmpp.NewStreamManager(client, nil)
	go func() {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Debugf("start XMPP listener for agent [%s]", a.LoginName)
		err = a.streamManagerService.Run()
		if err != nil {
			a.streamManagerService = nil
			log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Errorf("agent [%s] XMPP stream manager problem %s", a.LoginName, err)
			return
		}
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Tracef("started notify StreamManager for agent [%s]", a.LoginName)
		// await for stop
		<-a.ctx.Done()
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Tracef("stop notify subroutine for agent [%s]", a.LoginName)
		if a.streamManagerService != nil {
			a.streamManagerService.Stop()
			a.streamManagerService = nil
		}
	}()
	// wait for connect
	time.Sleep(1 * time.Second)
	return nil
}

// GetLastStatus get latest collected user status
func (a *Agent) GetLastStatus() *XmppUser {
	return a.lastStatus
}

// GetStatus geta actual agent status from finesse server
func (a *Agent) GetStatus() (*XmppUser, error) {
	request := a.newAgentRequest()
	response := request.doRequest("GET", a.server.urlString(request.id, "User", a.LoginName), nil)
	defer response.close()
	msg, err := response.responseError()
	if err != nil {
		log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Error(msg)
		return nil, err
	}
	log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Trace("success get data for agentId from name")
	data, err := newXmppUser(response.GetResponseBody())
	if err != nil {
		log.WithFields(log.Fields{logProc: "getAgentId", logId: response.id, logAgent: a.LoginName}).Error(err)
		return nil, err
	}
	a.lastStatus = data
	return a.lastStatus, nil
}

func (a *Agent) FullString() string {
	l := "Agent:"
	l = fmt.Sprintf("%s\r\n  Name:      %s", l, a.LoginName)
	l = fmt.Sprintf("%s\r\n  Id:        %s", l, a.LoginId)
	l = fmt.Sprintf("%s\r\n  Line:      %s", l, a.Line)
	if a.lastStatus != nil {
		l = fmt.Sprintf("%s\r\n  Status:      %s", l, a.lastStatus.State)
	} else {
		l = fmt.Sprintf("%s\r\n  Status:      %s", l, "UNKNOWN")
	}
	return l
}

func (a *Agent) String() string {
	if a.lastStatus != nil {
		return fmt.Sprintf("%s (%s) => %s", a.LoginName, a.Line, a.lastStatus.State)
	}
	return fmt.Sprintf("%s (%s) => UNKNOWN", a.LoginName, a.Line)
}

func (a *Agent) analyzeResponse(data string) (*XmppUser, error) {
	var envelope XmppUpdate
	var err error
	err = xml.Unmarshal([]byte(data), &envelope)
	if err != nil {
		log.WithFields(log.Fields{logProc: "analyzeResponse", logAgent: a.LoginName}).Warnf("problem with XML unmarshal envelope - %s", err)
		return nil, fmt.Errorf("problem unmarhal response XML")
	}
	if envelope.Data.Error.ApiErrors != nil {
		// problem here is error
		log.WithFields(log.Fields{logProc: "analyzeResponse", logAgent: a.LoginName}).Warnf("request ends with error for XMPP User - %s", err)
		return nil, fmt.Errorf("%s", envelope.Data.Error.ApiErrors[0].ErrorMessage)
	}
	if len(envelope.Data.User.URI) > 0 {
		log.WithFields(log.Fields{logProc: "analyzeResponse", logAgent: a.LoginName}).Tracef("collect User data for XMPP")
		usr := &envelope.Data.User
		return usr, nil
	}
	if envelope.Data.Dialogs.Dialogs != nil {
		log.WithFields(log.Fields{logProc: "analyzeResponse", logAgent: a.LoginName}).Warnf("collect dialogs data for XMPP User - %s", err)

	}
	log.WithFields(log.Fields{logProc: "analyzeResponse", logAgent: a.LoginName}).Warnf("unknown data body in request for XMPP User - %s", err)
	return nil, fmt.Errorf("problem get right data from response for request %s", envelope.Source)
}
