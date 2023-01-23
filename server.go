package finesse_api

import (
	"context"
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strings"
	"time"
)

// Server Structure for finesse server data
type Server struct {
	name         string // name is FQDN of server or IP address
	port         int    // port for finesse API
	ignore       bool   // ignore invalid certificate
	xmppPort     int    // port for XMPP notification
	insecureXmpp bool   // insecureXmpp for connect insecure direct XMPP instead of WSS
	timeOut      int    // timeOut for API requests default is 30 sec
}

const (
	DefaultServerHttpsPort      = 8445 // DefaultServerHttpsPort standard Finesse API port
	DefaultServerXmppPort       = 7443 // DefaultServerXmppPort standard secure XMPP over WSS port (secure XMPP communication)
	DefaultServerDirectXmppPort = 5222 // DefaultServerDirectXmppPort insecure XMPP port for direct communication (by default disabled on Finesse server)
	DefaultServerTimeout        = 30   // DefaultServerTimeout define timeout for API and Notify communication in seconds
)

// NewServer Creating new Finesse server structure connect on standard ports and manage if ignore certificate problems
//
// possible use repeat for different agents
func NewServer(server string, ignoreCert bool, time ...int) *Server {
	timeOut := DefaultServerTimeout
	if len(time) > 0 {
		timeOut = time[0]
	}
	return NewServerDetail(server, DefaultServerHttpsPort, ignoreCert, DefaultServerXmppPort, false, timeOut)
}

// NewServerDetail Create new Finesse server structure with required security and ports
//
// possible use repeat for different agents
func NewServerDetail(server string, port int, ignore bool, xmppPort int, insecureXmpp bool, timeOut int) *Server {
	return &Server{
		name:         server,
		port:         port,
		ignore:       ignore,
		xmppPort:     xmppPort,
		insecureXmpp: insecureXmpp,
		timeOut:      timeOut,
	}
}

// CreateAgent create new agent, read agent ID from Finesse server, start XMPP notify connection and return Agent or error if problem
//
//   - ctx context.Context - used for graceful shutdown of XMPP connection
func (s *Server) CreateAgent(ctx context.Context, name string, pwd string, line string) (*Agent, error) {
	log.WithFields(log.Fields{logProc: "AddAgent", logAgent: name}).Tracef("prepare agent and try collect it's ID")
	a := NewAgentNotify(ctx, name, pwd, line, s)
	err := a.getId()
	if err != nil {
		log.WithFields(log.Fields{logProc: "AddAgent", logAgent: name}).Tracef("can't get actual agent state")
		return nil, err
	}

	return a, nil
}

// getHttpClient create httpclient with setup from server configuration
func (s *Server) getHttpClient() *http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: s.ignore}
	client := &http.Client{Transport: customTransport, Timeout: time.Duration(s.timeOut) * time.Second}

	return client
}

// urlString create full API request path
//
// Expect:
//   - rId string - unique request identification
//   - pathPart ...string - one or more
//
// Example:
//   - urlString("xmp", "6350") => https://{server:port}/finesse/api/6350
//   - urlString("xmp", "6350", "Dialogs) => https://{server:port}/finesse/api/6350/Dialogs
func (s *Server) urlString(rId string, pathPart ...string) string {
	restPath := "/finesse/api"
	if len(pathPart) > 0 {
		for _, s := range pathPart {
			restPath = path.Join(restPath, s)
		}
	}
	var url string
	if s.port != 80 {
		url = fmt.Sprintf("https://%s:%d%s", s.name, s.port, restPath)
	} else {
		url = fmt.Sprintf("https://%s%s", s.name, restPath)
	}
	log.WithFields(log.Fields{logProc: "urlString", logServer: s.name, logId: rId}).Tracef("Request URI: %s", url)
	return url
}

// getDomain get only domain from Finesse server FQDN, for IP address or only host returns empty string
func (s *Server) getDomain() string {
	if validIpAddress(s.name) {
		return ""
	}
	l := strings.Split(s.name, ".")
	if len(l) > 1 {
		return strings.Join(l[1:], ".")
	}
	return ""
}
