package utilities

import (
	"fmt"
	"github.com/Spruik/libre-common/common/core/domain"
	"github.com/Spruik/libre-common/common/core/ports"
	libreConfig "github.com/Spruik/libre-configuration"
	libreLogger "github.com/Spruik/libre-logging"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type DaemonBase struct {
	libreConfig.ConfigurationEnabler
	libreLogger.LoggingEnabler

	name                  string
	state                 ports.DaemonStateIF
	initializationFxn     func(d ports.DaemonIF, params map[string]interface{}) error
	controlFxns           map[ports.DaemonCommandIF]ports.DaemonCommandFunction
	oneProcessingCycleFxn func(d ports.DaemonIF) (int, error)
	cleanupFxn            func(d ports.DaemonIF, params map[string]interface{}) error
	daemonChildren        []ports.DaemonIF
	parentWaitGroup       *sync.WaitGroup
	localWaitGroup        *sync.WaitGroup
	adminChannel          chan ports.DaemonAdminCommand
	terminationWaitGroup  *sync.WaitGroup
}

func NewDaemonBase(name string, initialState ports.DaemonStateIF, parentWG *sync.WaitGroup, configHook string) *DaemonBase {
	d := DaemonBase{}
	d.SetConfigCategory(configHook)
	hook, err := d.GetConfigItemWithDefault("loggerHook", "DAEMON")
	if err == nil {
		d.SetLoggerConfigHook(hook)
	} else {
		panic(fmt.Sprintf("Configuration Failure in NewDaemon: %s", err))
	}
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
	d.localWaitGroup = &sync.WaitGroup{}
	d.adminChannel = make(chan ports.DaemonAdminCommand)
	d.terminationWaitGroup = nil
	return &d
}

func (d *DaemonBase) Run(params map[string]interface{}) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		log.Printf("Daemon recieved system signal: %+v", s)
		//try to shutdown gracefully
		_, err := d.SubmitCommand(DaemonEndCommand, nil)
		if err != nil {
			log.Println("Failed graceful shutdown after system signal!")
		}
	}()
	go func() {
		if d.terminationWaitGroup != nil {
			d.terminationWaitGroup.Add(1)
			defer func() { d.terminationWaitGroup.Done() }()
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
			d.LogDebug(d.name, "calling acceptCommand while in state=", d.state)
			var cmdCycleEffect domain.DaemonCommandCycleEffect
			cmdCycleEffect, err = d.acceptCommand()
			if err != nil {
				panic(err)
			}
			switch cmdCycleEffect {
			case domain.DAEMON_CMD_EFFECT_CYCLE:
				run = true
			case domain.DAEMON_CMD_EFFECT_NOCYCLE:
				run = false
			case domain.DAEMON_CMD_EFFECT_UNCHANGED:
			}
			d.LogDebug(d.name, "consdering a cycle ", run, d.state)
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
		d.LogDebug(d.name, "calling cleanup while in state=", d.state)
		err = d.cleanupFxn(d, params)
		if err != nil {
			panic(err)
		}
	}()
}

func (d *DaemonBase) acceptCommand() (domain.DaemonCommandCycleEffect, error) {
	var cycleEffect domain.DaemonCommandCycleEffect
	var err error
	d.LogDebug(d.name, "checking for command")
	select {
	case chgCmd := <-d.adminChannel:
		d.LogDebug(d.name, "got command", chgCmd.Cmd)
		if d.parentWaitGroup != nil {
			d.LogDebug(d.name, "incrementing parent waitgroup", d.parentWaitGroup)
			d.parentWaitGroup.Add(1)
		}
		var resp = make(map[string]interface{})
		cmdFxn := d.controlFxns[chgCmd.Cmd]
		if cmdFxn != nil {
			resp, cycleEffect, err = cmdFxn(d, chgCmd.Params)
			if err != nil {
				panic(err)
			}
			d.LogDebug(d.name, "processed command message", chgCmd.Cmd, cycleEffect)
			if len(d.daemonChildren) > 0 {
				for _, child := range d.daemonChildren {
					d.LogDebug(d.name, "sending command to child", child.GetName(), chgCmd.Cmd)
					childresp, err := child.SubmitCommand(chgCmd.Cmd, chgCmd.Params)
					if err != nil {
						panic(err)
					}
					if childresp != nil {
						for key, val := range childresp {
							resp[key] = val
						}
					}
					d.LogDebug(d.name, "sent command message to child", child.GetName(), chgCmd.Cmd, cycleEffect)
				}
				d.LogDebug(d.name, "waiting for child completion", chgCmd.Cmd)
				d.localWaitGroup.Wait()
			}
			if chgCmd.Cmd.HasTargetState() {
				d.LogDebug("SETTING TARGET STATE TO ", chgCmd.Cmd.GetTargetState())
				d.SetState(chgCmd.Cmd.GetTargetState())
			}
			if len(resp) > 0 {
				chgCmd.Results = resp
			} else {
				chgCmd.Results = nil
			}
		} else {
			chgCmd.Results = map[string]interface{}{d.GetName(): "Command not implemented"}
		}
		d.adminChannel <- chgCmd
		if d.parentWaitGroup != nil {
			d.LogDebug(d.name, "telling parent wait group it's done ")
			d.parentWaitGroup.Done()
		}
	default:
		cycleEffect = domain.DAEMON_CMD_EFFECT_UNCHANGED
	}
	d.LogDebug(d.name, "done checking for command ", cycleEffect, err)
	return cycleEffect, err
}

func (d *DaemonBase) SetState(state ports.DaemonStateIF) {
	d.state = state
}
func (d *DaemonBase) GetState() ports.DaemonStateIF {
	return d.state
}
func (d *DaemonBase) ExecuteCommandFxn(cmd ports.DaemonCommandIF, params map[string]interface{}) (map[string]interface{}, domain.DaemonCommandCycleEffect, error) {
	fxn := d.controlFxns[cmd]
	if fxn != nil {
		return fxn(d, params)
	} else {
		panic("state change function not found")
	}
	return nil, domain.DAEMON_CMD_EFFECT_UNCHANGED, nil
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
	d.controlFxns[cmd] = fxn
}
func (d *DaemonBase) RemoveCommandFxn(cmd ports.DaemonCommandIF) {
	d.controlFxns[cmd] = nil
}
func (d *DaemonBase) AddDaemonChild(DaemonChild ports.DaemonIF) {
	DaemonChild.SetWaitGroup(d.localWaitGroup)
	d.daemonChildren = append(d.daemonChildren, DaemonChild)
}
func (d *DaemonBase) SubmitCommand(cmd ports.DaemonCommandIF, params map[string]interface{}) (map[string]interface{}, error) {
	d.adminChannel <- ports.DaemonAdminCommand{
		Cmd:     cmd,
		Params:  params,
		Results: nil,
		Err:     nil,
	}
	respObj := <-d.adminChannel
	return respObj.Results, respObj.Err
}
func (d *DaemonBase) SetTerminationWaitGroup(wg *sync.WaitGroup) {
	d.terminationWaitGroup = wg
}
func (d *DaemonBase) GetCommands() map[ports.DaemonCommandIF]ports.DaemonCommandFunction {
	return d.controlFxns
}

//////////////////////////////////////////////////////////////////////////////////////////
type DaemonCommand struct {
	name        string
	targetState ports.DaemonStateIF
}

func NewDaemonCommand(name string, state ports.DaemonStateIF) *DaemonCommand {
	return &DaemonCommand{name: name, targetState: state}
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

var DaemonRunCommand = NewDaemonCommand("Run", DaemonRunningState)
var DaemonPauseCommand = NewDaemonCommand("Pause", DaemonPausedState)
var DaemonEndCommand = NewDaemonCommand("End", DaemonEndState)
var DaemonGetStateCommand = NewDaemonCommand("GetState", nil)

////////////////////////////////////////////////////////////////////////////////////////////////////
type DaemonState struct {
	name       string
	canExecute bool
}

func NewDaemonState(name string, canExecute bool) *DaemonState {
	return &DaemonState{
		name:       name,
		canExecute: canExecute,
	}
}
func (s *DaemonState) GetStateName() string {
	return s.name
}
func (s *DaemonState) CanExecuteCycles() bool {
	return s.canExecute
}

var DaemonInitialState = NewDaemonState("INITIAL", false)
var DaemonRunningState = NewDaemonState("RUNNING", true)
var DaemonPausedState = NewDaemonState("PAUSED", false)
var DaemonEndState = NewDaemonState("ENDED", false)

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
func (s *standardFunctions) StandardRunFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, domain.DaemonCommandCycleEffect, error) {
	//do nothing
	return nil, domain.DAEMON_CMD_EFFECT_CYCLE, nil
}
func (s *standardFunctions) StandardEndFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, domain.DaemonCommandCycleEffect, error) {
	//do nothing
	return nil, domain.DAEMON_CMD_EFFECT_NOCYCLE, nil
}
func (s *standardFunctions) StandardPauseFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, domain.DaemonCommandCycleEffect, error) {
	//do nothing
	return nil, domain.DAEMON_CMD_EFFECT_CYCLE, nil
}
func (s *standardFunctions) GetStateFxn(d ports.DaemonIF, params map[string]interface{}) (map[string]interface{}, domain.DaemonCommandCycleEffect, error) {
	var resp = make(map[string]interface{})
	resp[d.GetName()] = d.GetState().GetStateName()
	return resp, domain.DAEMON_CMD_EFFECT_UNCHANGED, nil
}
