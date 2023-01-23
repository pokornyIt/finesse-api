package finesse_api

const (
	AgentStateLogin        string = "LOGIN"
	AgentStateLogout              = "LOGOUT"
	AgentStateReady               = "READY"
	AgentStateNotReady            = "NOT_READY"
	AgentStateAvailable           = "AVAILABLE"
	AgentStateTalking             = "TALKING"
	AgentStateWorkNotReady        = "WORK_NOT_READY"
	AgentStateWorkReady           = "WORK_READY"
	AgentStateReserved            = "RESERVED"
	AgentStateUnknown             = "UNKNOWN"
	AgentStateHold                = "HOLD"
	AgentStateActive              = "ACTIVE"
	AgentStatePaused              = "PAUSED"
	AgentStateInterrupted         = "INTERRUPTED"
	AgentStateNotActive           = "NOT_ACTIVE"
)

// AgentStates All valid agent states
var AgentStates = []string{AgentStateLogin, AgentStateLogout, AgentStateReady, AgentStateNotReady, AgentStateAvailable,
	AgentStateTalking, AgentStateWorkNotReady, AgentStateWorkReady, AgentStateReserved, AgentStateUnknown, AgentStateHold,
	AgentStateActive, AgentStatePaused, AgentStateInterrupted, AgentStateNotActive}

// AgentReadyStates States when agent is ready for work or work
var AgentReadyStates = map[string]string{AgentStateReady: AgentStateReady, AgentStateAvailable: AgentStateAvailable,
	AgentStateTalking: AgentStateTalking, AgentStateWorkReady: AgentStateWorkReady, AgentStateReserved: AgentStateReserved,
	AgentStateHold: AgentStateHold, AgentStateActive: AgentStateActive}

// AgentLoginStates States when agent is logged in to the system
var AgentLoginStates = map[string]string{AgentStateLogin: AgentStateLogin, AgentStateReady: AgentStateReady,
	AgentStateNotReady: AgentStateNotReady, AgentStateAvailable: AgentStateAvailable, AgentStateTalking: AgentStateTalking,
	AgentStateWorkNotReady: AgentStateWorkNotReady, AgentStateWorkReady: AgentStateWorkReady, AgentStateReserved: AgentStateReserved,
	AgentStateHold: AgentStateHold, AgentStateActive: AgentStateActive, AgentStatePaused: AgentStatePaused,
	AgentStateInterrupted: AgentStateInterrupted, AgentStateNotActive: AgentStateNotActive}

// AgentNotReadyStates States when agent is not-ready
var AgentNotReadyStates = map[string]string{AgentStateNotReady: AgentStateNotReady, AgentStateWorkNotReady: AgentStateWorkNotReady}

// AgentLogoutState States when agent is not logged in to the system
var AgentLogoutState = map[string]string{AgentStateLogout: AgentStateLogout, AgentStateUnknown: AgentStateUnknown}
