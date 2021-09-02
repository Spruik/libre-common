package utilities

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
)

type DaemonBase struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler

	name                   string
	state                  ports.DaemonStateIF
	initializationFxn      func(d ports.DaemonIF, params map[string]interface{}) error
	controlFxns            map[ports.DaemonCommandIF]ports.DaemonCommandFunction
	oneProcessingCycleFxn  func(d ports.DaemonIF) (int, error)
	cleanupFxn             func(d ports.DaemonIF, params map[string]interface{}) error
	daemonChildren         []ports.DaemonIF
	parentWaitGroup        *sync.WaitGroup
	localWaitGroup         sync.WaitGroup
	adminChannel           chan ports.DaemonAdminCommand
	terminationWaitGroup   *sync.WaitGroup
	commandMutex           sync.Mutex
	acceptCommandWaitLimit time.Duration
}

func NewDaemonBase(name string, initialState ports.DaemonStateIF, parentWG *sync.WaitGroup, configHook string) *DaemonBase {
	d := DaemonBase{}
	d.SetConfigCategory(configHook)
	loggerHook, cerr := d.GetConfigItemWithDefault(domain.LOGGER_CONFIG_HOOK_TOKEN, domain.DEFAULT_LOGGER_NAME)
	if cerr != nil {
		loggerHook = domain.DEFAULT_LOGGER_NAME
	}
	d.SetLoggerConfigHook(loggerHook)
	d.name = name
	d.parentWaitGroup = parentWG
	d.state = initialState
	stdFxns := NewStandardFunctions()
	d.initializationFxn = stdFxns.DaemonInitializeFunc
	controls := map[ports.DaemonCommandIF]ports.DaemonCommandFunction{}
	controls[DaemonRunCommand] = stdFxns.StandardRunFxn
	controls[DaemonEndCommand] = stdFxns.StandardEndFxn
	controls[DaemonPauseCommand] = stdFxns.StandardPauseFxn
	controls[DaemonGetStateCommand] = stdFxns.GetStateFxn
	d.controlFxns = controls
	d.oneProcessingCycleFxn = stdFxns.EmptyCycleFunc
	d.cleanupFxn = stdFxns.StandardCleanupFunc
	d.daemonChildren = make([]ports.DaemonIF, 0.0)
	d.localWaitGroup = sync.WaitGroup{}
	d.adminChannel = make(chan ports.DaemonAdminCommand)
	d.terminationWaitGroup = nil
	d.commandMutex = sync.Mutex{}
	acceptLimitStr, derr := d.GetConfigItemWithDefault("commandWaitDuration", "1000ms")
	if derr == nil {
		d.acceptCommandWaitLimit, derr = time.ParseDuration(acceptLimitStr)
	}
	if derr != nil {
		d.LogWarnf("Failed to configure Daemon '%s' commandWaitDuration - expecting a valid Go duration string such as '1000ms'.  Error=%s", d.name, derr)
	}
	return &d
}

func (d *DaemonBase) Run(params map[string]interface{}) {
	go func() {
		if d.terminationWaitGroup != nil {
			d.terminationWaitGroup.Add(1)
			defer func() {
				d.LogInfo("marking termination wait group done")
				d.terminationWaitGroup.Done()
			}()
			//if we have our termination wait group set, we must be the top level, so register the interrupts
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc,
				syscall.SIGHUP,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			go func() {
				s := <-sigc
				d.LogInfof("Notify recieved system signal: %+v", s)
				//try to shutdown gracefully
				d.LogInfof("sending END command to %s Daemon ", d.name)
				_, err := d.SubmitCommand(DaemonEndCommand, nil)
				if err != nil {
					d.LogInfof("%s Failed graceful shutdown after system signal!", d.name)
				}
				d.LogInfof("Done sending END command to %s Daemon ", d.name)
			}()

		}
		err := d.initializationFxn(d, params)
		if err != nil {
			panic(err)
		}
		for _, child := range d.daemonChildren {
			child.Run(params)
		}
		var run = true
		var processed = 0
		for run {
			d.LogDebug(d.name, "calling acceptCommand while in state=", d.state.GetStateName())
			err = d.acceptCommand()
			if err != nil {
				panic(err)
			}
			if d.state.IsTerminalState() {
				d.LogInfo(d.name, "latest command resulted in a terminal state - ending", d.state.GetStateName())
				run = false
			} else {
				d.LogDebug(d.name, "considering a cycle ", run, d.state.GetStateName())
				if run && d.state.CanExecuteCycles() {
					d.LogDebug(d.name, "starting a cycle")
					processed, err = d.oneProcessingCycleFxn(d)
					if err != nil {
						panic(err)
					}
					if processed == 0 {
						d.LogDebug(d.name, "Nothing processed this loop")
					}
				}
			}
		}
		d.LogDebug(d.name, "calling cleanup while in state=", d.state.GetStateName())
		err = d.cleanupFxn(d, params)
		if err != nil {
			panic(err)
		}
		d.LogInfof("%s run ends", d.name)
	}()
}

func (d *DaemonBase) acceptCommand() error {
	var err error
	d.LogDebug(d.name, "checking for command")
	select {
	case chgCmd := <-d.adminChannel:
		d.LogDebug(d.name, "got command", chgCmd.Cmd.GetCommandName())
		if d.parentWaitGroup != nil {
			d.LogDebug(d.name, "incrementing parent waitgroup")
			d.parentWaitGroup.Add(1)
			defer d.parentWaitGroup.Done()
		}
		resp := map[string]interface{}{}
		d.LogDebugf("%s looking for a function to implement %s where map is: %+v", d.name, chgCmd.Cmd.GetCommandName(), d.formatControlFxnMap())
		cmdFxn := d.controlFxns[chgCmd.Cmd]
		d.LogDebugf("%s looked for a function to implement %s where map is: %+v", d.name, chgCmd.Cmd.GetCommandName(), d.formatControlFxnMap())
		d.LogDebugf("%s found: %+v", d.name, cmdFxn)
		if cmdFxn != nil {
			resp, err = cmdFxn(d, chgCmd.Params)
			if err != nil {
				panic(err)
			}
			d.LogDebug(d.name, "processed command message", chgCmd.Cmd.GetCommandName())
		}
		if len(d.daemonChildren) > 0 {
			d.LogDebugf("%s start sending %s command to children", d.name, chgCmd.Cmd.GetCommandName())
			for _, child := range d.daemonChildren {
				d.LogDebug(d.name, "sending command to child", child.GetName(), chgCmd.Cmd.GetCommandName())
				childresp, err := child.SubmitCommand(chgCmd.Cmd, chgCmd.Params)
				if err != nil {
					panic(err)
				}
				if childresp != nil {
					resp[child.GetName()] = childresp
				}
				d.LogDebug(d.name, "sent command message to child", child.GetName(), chgCmd.Cmd.GetCommandName())
			}
			d.LogDebugf("%s waiting for child completion of %s", d.name, chgCmd.Cmd.GetCommandName())
			d.localWaitGroup.Wait()
			d.LogDebugf("%s done waiting for child completion of %s", d.name, chgCmd.Cmd.GetCommandName())
		}
		if chgCmd.Cmd.HasTargetState() {
			d.LogInfof("DAEMON '%s' SETTING STATE TO %s", d.name, chgCmd.Cmd.GetTargetState().GetStateName())
			d.SetState(chgCmd.Cmd.GetTargetState())
		}
		if len(resp) > 0 {
			chgCmd.Results = resp
		} else {
			chgCmd.Results = nil
		}
		d.adminChannel <- chgCmd
	case <-time.After(d.acceptCommandWaitLimit):
		d.LogDebugf("%s waited %+v for a command and did not receive one", d.name, d.acceptCommandWaitLimit)
	}
	d.LogDebug(d.name, "done checking for command.  err=%+v ", err)
	return err
}

func (d *DaemonBase) formatControlFxnMap() string {
	var ret = ""
	for key, val := range d.controlFxns {
		ret += fmt.Sprintf("%s:%+v ", key.GetCommandName(), val)
	}
	return ret
}
func (d *DaemonBase) SetState(state ports.DaemonStateIF) {
	d.state = state
}
func (d *DaemonBase) GetState() ports.DaemonStateIF {
	return d.state
}
func (d *DaemonBase) ExecuteCommandFxn(cmd ports.DaemonCommandIF, params map[string]interface{}) (map[string]interface{}, error) {
	fxn := d.controlFxns[cmd]
	if fxn != nil {
		return fxn(d, params)
	}
	panic("state change function not found")

}
func (d *DaemonBase) GetName() string {
	return d.name
}
func (d *DaemonBase) GetAdminChannel() chan ports.DaemonAdminCommand {
	return d.adminChannel
}
func (d *DaemonBase) SetWaitGroup(wg *sync.WaitGroup) {
	d.parentWaitGroup = wg
}
func (d *DaemonBase) SetInitializationFxn(fxn func(d ports.DaemonIF, params map[string]interface{}) error) {
	d.initializationFxn = fxn
}
func (d *DaemonBase) SetOneProcessingCycleFxn(fxn func(d ports.DaemonIF) (int, error)) {
	d.oneProcessingCycleFxn = fxn
}
func (d *DaemonBase) SetCleanupFxn(fxn func(d ports.DaemonIF, params map[string]interface{}) error) {
	d.cleanupFxn = fxn
}
func (d *DaemonBase) AddCommandFxn(cmd ports.DaemonCommandIF, fxn ports.DaemonCommandFunction) {
	d.LogInfof("%s ADDING COMMAND FUNCTION FOR %s", d.name, cmd.GetCommandName())
	d.controlFxns[cmd] = fxn
}
func (d *DaemonBase) RemoveCommandFxn(cmd ports.DaemonCommandIF) {
	d.controlFxns[cmd] = nil
}
func (d *DaemonBase) AddDaemonChild(DaemonChild ports.DaemonIF) {
	DaemonChild.SetWaitGroup(&d.localWaitGroup)
	d.daemonChildren = append(d.daemonChildren, DaemonChild)
}
func (d *DaemonBase) RemoveDaemonChild(DaemonChild ports.DaemonIF) {
	_, err := DaemonChild.SubmitCommand(DaemonEndCommand, nil)
	if err != nil {
		d.LogErrorf("%s FAILED END COMMAND DURING REMOVE OF CHILD %s", d.name, DaemonChild.GetName())
	}
	var delndx int = -1
	for ndx := 0; ndx < len(d.daemonChildren); ndx++ {
		if d.daemonChildren[ndx] == DaemonChild {
			delndx = ndx
			break
		}
	}
	if delndx != -1 {
		d.daemonChildren = append(d.daemonChildren[:delndx], d.daemonChildren[delndx+1:]...)
	}
}
func (d *DaemonBase) SubmitCommand(cmd ports.DaemonCommandIF, params map[string]interface{}) (map[string]interface{}, error) {
	d.LogDebugf("SubmitCommand called to send %s to %s ", cmd.GetCommandName(), d.name)
	d.commandMutex.Lock()
	defer d.commandMutex.Unlock()
	d.adminChannel <- ports.DaemonAdminCommand{
		Cmd:     cmd,
		Params:  params,
		Results: nil,
		Err:     nil,
	}
	d.LogDebugf("SubmitCommand waiting for response to %s from %s", cmd.GetCommandName(), d.name)
	respObj := <-d.adminChannel
	d.LogDebugf("SubmitCommand received response to %s from %s", cmd.GetCommandName(), d.name)
	//return map[string]interface{}{d.name: respObj.Results}, respObj.Err
	return respObj.Results, respObj.Err
}
func (d *DaemonBase) SetTerminationWaitGroup(wg *sync.WaitGroup) {
	d.terminationWaitGroup = wg
}
func (d *DaemonBase) GetCommands() map[ports.DaemonCommandIF]ports.DaemonCommandFunction {
	ret := d.controlFxns
	for _, dchild := range d.daemonChildren {
		for i, j := range dchild.GetCommands() {
			ret[i] = j
		}
	}
	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////
type DaemonCommand struct {
	name        string
	targetState ports.DaemonStateIF
	inParams    []string
}

func NewDaemonCommand(name string, state ports.DaemonStateIF, inParams []string) *DaemonCommand {
	return &DaemonCommand{name: name, targetState: state, inParams: inParams}
}
func (s *DaemonCommand) GetCommandName() string {
	return s.name
}
func (s *DaemonCommand) HasTargetState() bool {
	return !(s.targetState == nil)
}
func (s *DaemonCommand) GetTargetState() ports.DaemonStateIF {
	if s.targetState != nil {
		return s.targetState
	} else {
		panic("GetTargetState called when internal pointer is nil")
	}
}
func (s *DaemonCommand) GetInputParamNames() []string {
	return s.inParams
}

var DaemonRunCommand = NewDaemonCommand("Run", DaemonRunningState, nil)
var DaemonPauseCommand = NewDaemonCommand("Pause", DaemonPausedState, nil)
var DaemonEndCommand = NewDaemonCommand("End", DaemonEndState, nil)
var DaemonGetStateCommand = NewDaemonCommand("GetState", nil, nil)

////////////////////////////////////////////////////////////////////////////////////////////////////
type DaemonState struct {
	name       string
	canExecute bool
	isTerminal bool
}

func NewDaemonState(name string, canExecute bool, isTerminal bool) *DaemonState {
	return &DaemonState{
		name:       name,
		canExecute: canExecute,
		isTerminal: isTerminal,
	}
}
func (s *DaemonState) GetStateName() string {
	return s.name
}
func (s *DaemonState) CanExecuteCycles() bool {
	return s.canExecute
}
func (s *DaemonState) IsTerminalState() bool {
	return s.isTerminal
}

var DaemonInitialState = NewDaemonState("INITIAL", false, false)
var DaemonRunningState = NewDaemonState("RUNNING", true, false)
var DaemonPausedState = NewDaemonState("PAUSED", false, false)
var DaemonEndState = NewDaemonState("ENDED", false, true)

//////////////////////////////////////////////////////////////////////////////////////////////////////////
type standardFunctions struct {
}

func NewStandardFunctions() *standardFunctions {
	s := standardFunctions{}
	return &s
}

func (s *standardFunctions) DaemonInitializeFunc(d ports.DaemonIF, params map[string]interface{}) error {
	//do nothing
	return nil
}
func (s *standardFunctions) StandardCleanupFunc(d ports.DaemonIF, params map[string]interface{}) error {
	//do nothing
	return nil
}
func (s *standardFunctions) EmptyCycleFunc(d ports.DaemonIF) (int, error) {
	//do nothing
	return 0, nil
}
func (s *standardFunctions) StandardRunFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, error) {
	//do nothing
	return map[string]interface{}{}, nil
}
func (s *standardFunctions) StandardEndFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, error) {
	//do nothing
	return map[string]interface{}{}, nil
}
func (s *standardFunctions) StandardPauseFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, error) {
	//do nothing
	return map[string]interface{}{}, nil
}
func (s *standardFunctions) GetStateFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, error) {
	var resp = make(map[string]interface{})
	resp["State"] = d.GetState().GetStateName()
	return resp, nil
}
