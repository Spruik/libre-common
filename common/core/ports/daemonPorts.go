package ports

import (
	"sync"
)

type DaemonAdminCommand struct {
	Cmd     DaemonCommandIF
	Params  map[string]interface{}
	Err     error
	Results map[string]interface{}
}

type DaemonCommandFunction func(d DaemonIF, params map[string]interface{}) (map[string]interface{}, error)

type DaemonIF interface {
	Run(params map[string]interface{})
	GetName() string
	GetAdminChannel() chan DaemonAdminCommand
	GetState() DaemonStateIF
	SetState(state DaemonStateIF)
	SetWaitGroup(wg *sync.WaitGroup)
	SubmitCommand(cmd DaemonCommandIF, params map[string]interface{}) (map[string]interface{}, error)
	SetInitializationFxn(fxn func(d DaemonIF, params map[string]interface{}) error)
	SetOneProcessingCycleFxn(fxn func(d DaemonIF) (int, error))
	SetCleanupFxn(fxn func(d DaemonIF, params map[string]interface{}) error)
	AddCommandFxn(cmd DaemonCommandIF, fxn DaemonCommandFunction)
	RemoveCommandFxn(cmd DaemonCommandIF)
	AddDaemonChild(DaemonChild DaemonIF)
	RemoveDaemonChild(DaemonChild DaemonIF)
	SetTerminationWaitGroup(wg *sync.WaitGroup)
	GetCommands() map[DaemonCommandIF]DaemonCommandFunction
}

type DaemonStateIF interface {
	GetStateName() string
	CanExecuteCycles() bool
	IsTerminalState() bool
}

type DaemonCommandIF interface {
	GetCommandName() string
	HasTargetState() bool
	GetTargetState() DaemonStateIF
	GetInputParamNames() []string
}
