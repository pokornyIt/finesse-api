package finesse_api

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
)

// massive parallel processing sets agents into the same state

type AgentGroup struct {
	Agents     []*Agent
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
	mutex      sync.Mutex
}

type BulkAgent struct {
	Name     string
	Password string
	Line     string
}

func NewAgentGroup() *AgentGroup {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &AgentGroup{
		Agents:     nil,
		ctx:        ctx,
		cancelFunc: cancelFunc,
		wg:         sync.WaitGroup{},
	}
}

func (group *AgentGroup) AddAgentToGroup(name string, pwd string, line string, server *Server) error {
	agent := NewAgentNotify(group.ctx, name, pwd, line, server)
	err := agent.getId()
	if err != nil {
		log.WithFields(log.Fields{logProc: "AddAgentToGroup", logAgent: name}).Errorf("can't get actual agent state")
		return err
	}
	if err = agent.StartXmpp(); err != nil {
		log.WithFields(log.Fields{logProc: "AddAgentToGroup", logId: agent.LoginId, logServer: server.name}).
			Errorf("problem start XMPP for agent [%s] on server [%s]", agent.LoginName, server.name)
		return err
	}
	log.WithFields(log.Fields{logProc: "AddAgentToGroup", logId: agent.LoginId, logServer: server.name}).
		Tracef("start XMPP subroutine for agent [%s] on server [%s]", agent.LoginName, server.name)
	group.Agents = append(group.Agents, agent)
	return nil
}

func (group *AgentGroup) AddBulkAgents(agents []BulkAgent, server *Server) []OperationError {
	log.WithFields(log.Fields{logProc: "AddBulkAgents"}).Tracef("start procees add bulk agents with it's status")
	err := make(chan OperationError, len(agents))
	for _, agent := range agents {
		group.wg.Add(1)
		go func(a BulkAgent, server *Server, wg *sync.WaitGroup, c chan OperationError) {
			defer wg.Done()

			ag, e := server.CreateAgent(group.ctx, a.Name, a.Password, a.Line)
			if e != nil {
				c <- OperationError{
					Type:  TypeErrorNoStatus,
					Error: e,
				}
				return
			}
			e = ag.StartXmpp()
			if e != nil {
				c <- OperationError{
					Type:  TypeErrorNoStatus,
					Error: e,
				}
				return
			}
			group.mutex.Lock()
			group.Agents = append(group.Agents, ag)
			group.mutex.Unlock()
			c <- OperationError{
				Type:  TypeErrorNoError,
				Error: nil,
			}
		}(agent, server, &group.wg, err)
	}
	log.WithFields(log.Fields{logProc: "AddBulkAgents"}).Trace("wait for finish all in group")
	group.wg.Wait()
	var ret []OperationError
	log.WithFields(log.Fields{logProc: "AddBulkAgents"}).Trace("wait for finish all in group")
	for len(err) > 0 {
		ret = append(ret, <-err)
	}
	if len(ret) != len(agents) {
		log.WithFields(log.Fields{logProc: "AddBulkAgents"}).
			Errorf("Agents in AgentGroup [%d] different from number of responses [%d]", len(agents), len(err))
	} else {
		log.WithFields(log.Fields{logProc: "AddBulkAgents"}).
			Trace("operation processed for all Agent in group")
	}
	return ret
}

func (group *AgentGroup) Login() []OperationError {
	return group.doRequest(AgentStateLogin, false)
}

func (group *AgentGroup) Logout(forceLogout ...bool) []OperationError {
	force := false
	if len(forceLogout) > 0 {
		force = forceLogout[0]
	}
	return group.doRequest(AgentStateLogout, force)
}

func (group *AgentGroup) Ready(forceLogout ...bool) []OperationError {
	force := false
	if len(forceLogout) > 0 {
		force = forceLogout[0]
	}
	return group.doRequest(AgentStateReady, force)
}

func (group *AgentGroup) NotReady() []OperationError {
	return group.doRequest(AgentStateNotReady, false)
}

func (group *AgentGroup) CancelFunction() {
	if group.cancelFunc != nil {
		log.WithFields(log.Fields{logProc: "CancelFunction"}).
			Trace("call cancelFunc for all subroutines")
		group.cancelFunc()
	}
}

func (group *AgentGroup) doRequest(operation string, force bool) []OperationError {
	lProc := "doRequest"
	if group.Agents == nil || len(group.Agents) < 1 {
		log.WithFields(log.Fields{logProc: lProc, logRequestType: operation}).
			Warn("AgentGroup is empty")
		return nil
	}
	err := make(chan OperationError, len(group.Agents))
	for _, agent := range group.Agents {
		group.wg.Add(1)
		go agentOperation(operation, agent, &group.wg, err, force)
	}
	group.wg.Wait()
	var ret []OperationError
	for len(err) > 0 {
		ret = append(ret, <-err)
	}
	if len(ret) != len(group.Agents) {
		log.WithFields(log.Fields{logProc: lProc, logRequestType: operation}).
			Errorf("Agents in AgentGroup [%d] different from number of responses [%d]", len(group.Agents), len(err))
	} else {
		log.WithFields(log.Fields{logProc: lProc, logRequestType: operation}).
			Trace("operation processed for all Agent in group")
	}
	return ret
}

func agentOperation(operation string, a *Agent, wg *sync.WaitGroup, c chan OperationError, force bool) {
	lProc := "agentOperation"
	log.WithFields(log.Fields{logProc: lProc, logRequestType: operation, logAgent: a.LoginName}).
		Tracef("process operation [%s] for agent [%s]", operation, a.LoginName)
	defer wg.Done()
	switch operation {
	case AgentStateLogin:
		c <- a.Login()
	case AgentStateLogout:
		c <- a.Logout(force)
	case AgentStateReady:
		c <- a.Ready(force)
	case AgentStateNotReady:
		c <- a.NotReady()
	default:
		log.WithFields(log.Fields{logProc: lProc, logRequestType: operation, logAgent: a.LoginName}).
			Errorf("unknown operation [%s] for agent [%s]", operation, a.LoginName)
		c <- OperationError{
			Type:  TypeErrorUnknownBulkCommand,
			Error: fmt.Errorf("unknown AgentGroup operation [%s] for agent [%s]", operation, a.LoginName),
		}
	}
}
