//go:build ignore

package finesse_api

//
// this package is prepared for next use when identify how really work with Notification Service XMPP
//
import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gosrc.io/xmpp"
	"gosrc.io/xmpp/stanza"
	"os"
	"sync"
	"time"
)

// FinesseAgentNotify Structure for agent state with wait for XMPP response
type FinesseAgentNotify struct {
	*FinesseAgent
	XmppClient *xmpp.Client
	UserDetail *UserDetailResponse
	Mutex      sync.Mutex
}

// NewAgentWithNotification Create necessary agent notify structure
func NewAgentWithNotification(id string, name string, pwd string, line string) *FinesseAgentNotify {
	a := NewAgent(id, name, pwd, line)
	return &FinesseAgentNotify{
		FinesseAgent: a,
		XmppClient:   nil,
		UserDetail:   nil,
		Mutex:        sync.Mutex{},
	}
}

// StartNotification Test procedure for use XMPP
func (a *FinesseAgentNotify) StartNotification(finesseServer string) error {
	if a.XmppClient != nil {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("start finesse_notifier - XMPP notifier is ready")
	}
	log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("start finesse_notifier")
	config := xmpp.Config{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: fmt.Sprintf("%s:5223", finesseServer),
		},
		Jid:          a.LoginName,
		Credential:   xmpp.Password(a.Password),
		StreamLogger: os.Stdout,
		Insecure:     true,
	}

	router := xmpp.NewRouter()
	router.HandleFunc("iq", func(s xmpp.Sender, p stanza.Packet) {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("handle func iq")
	})
	router.HandleFunc("message", func(s xmpp.Sender, p stanza.Packet) {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("handle func message")
	})

	client, err := xmpp.NewClient(&config, router, func(err error) {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Errorf("client error handler messages - %s", err)
	})
	log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Trace("prepare XMPP client")

	err = client.Connect()
	if err != nil {
		log.WithFields(log.Fields{logProc: "StartNotification", logAgent: a.LoginName}).Errorf("XMPP client not connect %s", err)
		return err
	}
	// wait until response
	time.Sleep(SleepDurationMilliSecond)
	a.XmppClient = client
	return nil
}
