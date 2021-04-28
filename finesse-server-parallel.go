package finesse_api

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// ResponseChan Structure for collect responses in parallel functions
type ResponseChan struct {
	agentResponse *UserDetailResponse
	requestError  error
}

// ToString Return formatted string with responses for agents.
// The format is Name, Current state, Pending State or Name, Current state, Error
func (r *ResponseChan) ToString() string {
	if r.requestError != nil {
		return fmt.Sprintf("Agent: %-30s State: %-15s Error: %s", r.agentResponse.LoginName, r.agentResponse.State, r.requestError)
	}
	return fmt.Sprintf("Agent: %-30s State: %-15s Pending state: %s", r.agentResponse.LoginName, r.agentResponse.State, r.agentResponse.PendingState)
}

// LoginAllParallel Logging in all agents registered in the server.
// Every agent has own goroutine
func (f *FinesseServer) LoginAllParallel() error {
	fn := f.loginAgentRoutine
	return f.parallelProcessing(fn, "LoginAllParallel", "login")
}

// ReadyAllAgentsParallel Setting all agents registered in the server to state ready.
// Every agent has own goroutine
func (f *FinesseServer) ReadyAllAgentsParallel() error {
	fn := f.readyAgentRoutine
	return f.parallelProcessing(fn, "ReadyAllAgentsParallel", "ready")
}

func (f *FinesseServer) loginAgentRoutine(agentName string, c chan bool, wg *sync.WaitGroup) {
	log.WithFields(log.Fields{logProc: "loginAgentRoutine", logAgent: agentName}).Tracef("try login agent %s in separate thread", agentName)
	c <- f.LoginAgent(agentName)
	time.Sleep(SleepDurationMilliSecond)
	wg.Done()
}

func (f *FinesseServer) readyAgentRoutine(agentName string, c chan bool, wg *sync.WaitGroup) {
	c <- f.ReadyAgent(agentName)
	time.Sleep(SleepDurationMilliSecond)
	wg.Done()
}

func (f *FinesseServer) parallelProcessing(fn func(agentName string, c chan bool, wg *sync.WaitGroup), mainName string, operation string) error {
	ts := time.Now()

	defer logEndTimeProcess(ts, log.Fields{logProc: mainName, logServer: f.name})
	c := make(chan bool, len(f.agents))
	var wg sync.WaitGroup
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("start coroutines for %s", operation)
	for _, agent := range f.agents {
		wg.Add(1)
		go fn(agent.LoginName, c, &wg)
	}
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("wait for all coroutines %s ends", operation)
	wg.Wait()
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("all coroutines %s ends", operation)
	success := 0
	for range f.agents {
		if <-c {
			success++
		}
	}
	close(c)

	if len(f.agents) > success {
		log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Errorf("from %d agents success requests is %d", len(f.agents), success)
		return fmt.Errorf("%d problem agents in %s", len(f.agents)-success, operation)
	}
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Infof("from %d agents success requests is %d", len(f.agents), success)
	return nil
}

// GetStateAgentsParallel Getting current statuses for all agents registered in the server.
// Every agent has own goroutine
func (f *FinesseServer) GetStateAgentsParallel() ([]*UserDetailResponse, error) {
	fn := f.stateAgentRoutineWithStatus
	return f.parallelProcessingWithState(fn, false, "GetStateAgentsParallel", "status")
}

// LoginAgentsParallelWithStatus Logging in and getting current statuses for all agents registered in the server.
// Every agent has own goroutine
func (f *FinesseServer) LoginAgentsParallelWithStatus() ([]*UserDetailResponse, error) {
	fn := f.loginAgentRoutineWithStatus
	return f.parallelProcessingWithState(fn, false, "LoginAgentsParallelWithStatus", "login")
}

// ReadyAgentsParallelWithStatus Setting ready and getting current statuses for all agents registered in the server.
// The optional force parameter defines if the program tries logging in not logged-in agents, before setts it to a ready state
// Every agent has own goroutine
func (f *FinesseServer) ReadyAgentsParallelWithStatus(force ...bool) ([]*UserDetailResponse, error) {
	fo := false
	if len(force) > 0 {
		fo = force[0]
	}
	fn := f.readyAgentRoutineWithStatus
	return f.parallelProcessingWithState(fn, fo, "ReadyAllAgentsParallelWithStatus", "ready")
}

// NotReadyAgentsParallelWithStatus Setting not-ready and getting current statuses for all agents registered in the server.
// The optional force parameter defines if the program tries logging in not logged-in agents, before setts it to a not-ready state
// Every agent has own goroutine
func (f *FinesseServer) NotReadyAgentsParallelWithStatus(force ...bool) ([]*UserDetailResponse, error) {
	fo := false
	if len(force) > 0 {
		fo = force[0]
	}
	fn := f.notReadyAgentRoutineWithStatus
	return f.parallelProcessingWithState(fn, fo, "NotReadyAgentsParallelWithStatus", "not-ready")
}

// LogoutAgentsParallelWithStatus Logging out and getting current statuses for all agents registered in the server.
// The optional force parameter defines if the program tries sets a not-ready state before logging out
// Every agent has own goroutine
func (f *FinesseServer) LogoutAgentsParallelWithStatus(force ...bool) ([]*UserDetailResponse, error) {
	fo := false
	if len(force) > 0 {
		fo = force[0]
	}
	fn := f.logoutAgentRoutineWithStatus
	return f.parallelProcessingWithState(fn, fo, "LogoutAgentsParallelWithStatus", "logout")
}

func (f *FinesseServer) stateAgentRoutineWithStatus(agentName string, _ bool, c chan ResponseChan, wg *sync.WaitGroup) {
	status, err := f.GetAgentStatusDetail(agentName)
	defer wg.Done()
	if err != nil {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  err,
		}
	} else {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
	}
}

func (f *FinesseServer) loginAgentRoutineWithStatus(agentName string, _ bool, c chan ResponseChan, wg *sync.WaitGroup) {
	status, err := f.GetAgentStatusDetail(agentName)
	defer wg.Done()
	if err != nil {
		c <- ResponseChan{
			agentResponse: nil,
			requestError:  err,
		}
		return
	}
	if status.State == AgentStateUnknown {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  fmt.Errorf("agent [%s] is not in state for set to ready, actual state [%s]", agentName, status.State),
		}
		return
	}
	_, exists := AgentLoginStates[status.State]
	if exists {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	_ = f.LoginAgent(agentName)
	time.Sleep(SleepDurationMilliSecond)
	status, err = f.GetAgentStatusDetail(agentName)
	_, exists = AgentLoginStates[status.State]
	if exists {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	c <- ResponseChan{
		agentResponse: status,
		requestError:  fmt.Errorf("problem login agent [%s] to system", agentName),
	}
}

func (f *FinesseServer) readyAgentRoutineWithStatus(agentName string, force bool, c chan ResponseChan, wg *sync.WaitGroup) {
	status, err := f.GetAgentStatusDetail(agentName)
	defer wg.Done()
	if err != nil {
		c <- ResponseChan{
			agentResponse: nil,
			requestError:  err,
		}
		return
	}
	if !force && (status.State == AgentStateUnknown || status.State == AgentStateLogout) {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  fmt.Errorf("agent [%s] is not in state for set to ready, actual state [%s]", agentName, status.State),
		}
		return
	}
	if force && status.State == AgentStateLogout {
		_ = f.LoginAgent(agentName)
		time.Sleep(SleepDurationMilliSecond)
	}
	ready := f.ReadyAgent(agentName)
	time.Sleep(SleepDurationMilliSecond)
	status, err = f.GetAgentStatusDetail(agentName)
	_, exists := AgentReadyStates[status.State]
	if ready && exists {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	c <- ResponseChan{
		agentResponse: status,
		requestError:  fmt.Errorf("problem set agent [%s] to ready state", agentName),
	}
}

func (f *FinesseServer) notReadyAgentRoutineWithStatus(agentName string, force bool, c chan ResponseChan, wg *sync.WaitGroup) {
	status, err := f.GetAgentStatusDetail(agentName)
	defer wg.Done()
	if err != nil {
		c <- ResponseChan{
			agentResponse: nil,
			requestError:  err,
		}
		return
	}
	if !force && status.State == AgentStateUnknown {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  fmt.Errorf("agent [%s] is not in state for set to ready, actual state [%s]", agentName, status.State),
		}
		return
	}
	if status.State == AgentStateLogout {
		_ = f.LoginAgent(agentName)
	} else {
		_, exists := AgentReadyStates[status.State]
		if force && exists {
			_ = f.NotReadyAgent(agentName)
		}
	}
	time.Sleep(SleepDurationMilliSecond)
	status, err = f.GetAgentStatusDetail(agentName)

	_, exists := AgentNotReadyStates[status.State]
	_, exists1 := AgentNotReadyStates[status.PendingState]
	if exists || exists1 {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	c <- ResponseChan{
		agentResponse: status,
		requestError:  fmt.Errorf("problem set agent [%s] to not-ready state", agentName),
	}
}

func (f *FinesseServer) logoutAgentRoutineWithStatus(agentName string, force bool, c chan ResponseChan, wg *sync.WaitGroup) {
	status, err := f.GetAgentStatusDetail(agentName)
	defer wg.Done()
	if err != nil {
		c <- ResponseChan{
			agentResponse: nil,
			requestError:  err,
		}
		return
	}
	if !force && status.State == AgentStateUnknown {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  fmt.Errorf("agent [%s] is not in state for set to ready, actual state [%s]", agentName, status.State),
		}
		return
	}
	if status.State == AgentStateLogout {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	_, exists := AgentReadyStates[status.State]
	if force && exists {
		nr := f.NotReadyAgent(agentName)
		if !nr {
			c <- ResponseChan{
				agentResponse: status,
				requestError:  fmt.Errorf("agent [%s] not possible switch to not ready state, actual state [%s]", agentName, status.State),
			}
			return
		}
		time.Sleep(SleepDurationMilliSecond)
		exists = false
	}
	if exists {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  fmt.Errorf("agent [%s] not possible switch to not ready state, actual state [%s]", agentName, status.State),
		}
		return
	}
	_ = f.LogoutAgent(agentName)
	time.Sleep(SleepDurationMilliSecond)
	status, err = f.GetAgentStatusDetail(agentName)

	if status.State == AgentStateLogout {
		c <- ResponseChan{
			agentResponse: status,
			requestError:  nil,
		}
		return
	}
	c <- ResponseChan{
		agentResponse: status,
		requestError:  fmt.Errorf("problem logout agent [%s]", agentName),
	}
}

func (f *FinesseServer) parallelProcessingWithState(fn func(agentName string, force bool, c chan ResponseChan, wg *sync.WaitGroup), force bool, mainName string, operation string) ([]*UserDetailResponse, error) {
	ts := time.Now()

	var status []*UserDetailResponse
	defer logEndTimeProcess(ts, log.Fields{logProc: mainName, logServer: f.name})
	c := make(chan ResponseChan, len(f.agents))
	var wg sync.WaitGroup
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("start coroutines for %s", operation)
	for _, agent := range f.agents {
		wg.Add(1)
		go fn(agent.LoginName, force, c, &wg)
	}
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("wait for all coroutines %s ends", operation)
	wg.Wait()
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Debugf("all coroutines %s ends", operation)
	success := 0
	for range f.agents {
		r := <-c
		status = append(status, r.agentResponse)
		if r.requestError == nil {
			success++
		}
	}
	close(c)

	if len(f.agents) > success {
		log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Errorf("from %d agents success requests is %d", len(f.agents), success)
		return status, fmt.Errorf("%d problem agents in %s", len(f.agents)-success, operation)
	}
	log.WithFields(log.Fields{logProc: mainName, logServer: f.name}).Infof("from %d agents success requests is %d", len(f.agents), success)
	return status, nil
}

func logEndTimeProcess(ts time.Time, fields log.Fields) {
	end := time.Now()
	log.WithFields(fields).WithField(runTimeDuration, end.Sub(ts).String()).Tracef("processing duration: %s", end.Sub(ts))
}
