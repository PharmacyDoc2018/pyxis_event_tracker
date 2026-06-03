package main

import "fmt"

type processMode int

const (
	Normal = iota
	LocalOnly
	SafetyMode
)

type processState struct {
	pyxisLogsLoaded   []string
	mode              processMode
	eRxItemIdLinks    bool
	dbConnection      bool
	pyxisEventsLoaded bool
}

func (p *processState) UpdateState() {
	if !p.eRxItemIdLinks {
		p.mode = SafetyMode
		return
	}

	if !p.pyxisEventsLoaded {
		p.mode = SafetyMode
		return
	}

	if !p.dbConnection {
		p.mode = LocalOnly
		return
	}
}

func (p *processState) Mode() processMode {
	return p.mode
}

func (p *processState) ERxItemIdLinksOkay() bool {
	return p.eRxItemIdLinks
}

func (p *processState) ERxItemIdLinksSuccessful() {
	p.eRxItemIdLinks = true
	p.UpdateState()
}

func (p *processState) DbConnectionOkay() bool {
	return p.dbConnection
}

func (p *processState) DbConnectionSuccessful() {
	p.dbConnection = true
	p.UpdateState()
}

func (p *processState) DbConnectionFail() {
	p.dbConnection = false
	p.UpdateState()
}

func (p *processState) PyxisEventLogsLoadedOkay() bool {
	return p.pyxisEventsLoaded
}

func (p *processState) PyxisEventLogsLoadSuccessful() {
	p.pyxisEventsLoaded = true
	p.UpdateState()
}

func (p *processState) PyxisEventLogLoaded(pyxis string) *logError {
	for _, log := range p.pyxisLogsLoaded {
		if log == pyxis {
			return &logError{
				errMessage: fmt.Sprintf("error. %s already in loaded state", pyxis),
				logMessage: fmt.Sprintf("Error. %s already in loaded state", pyxis),
			}
		}
	}

	p.pyxisLogsLoaded = append(p.pyxisLogsLoaded, pyxis)
	return nil
}

func (p *processState) PyxisEventLogUnloaded(pyxis string) *logError {
	for i, log := range p.pyxisLogsLoaded {
		if log == pyxis {
			p.pyxisLogsLoaded = append(p.pyxisLogsLoaded[:i], p.pyxisLogsLoaded[i+1:]...)
			return nil
		}
	}

	return &logError{
		errMessage: fmt.Sprintf("error. %s not in loaded state", pyxis),
		logMessage: fmt.Sprintf("Error. %s not in loaded state", pyxis),
	}
}

func (p *processState) IsLoaded(pyxis string) bool {
	for _, log := range p.pyxisLogsLoaded {
		if log == pyxis {
			return true
		}
	}

	return false
}

func initProcessState() *processState {
	return &processState{
		mode: SafetyMode,
	}
}
