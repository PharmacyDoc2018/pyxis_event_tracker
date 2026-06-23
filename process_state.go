package main

import (
	"fmt"
	"strconv"
)

type processMode int

const (
	Normal = iota
	LocalOnly
	SafetyMode
	TestMode
)

type processState struct {
	pyxisLogsLoaded    []string
	mode               processMode
	erxs               bool
	eRxItemIdLinks     bool
	departmentCoverage bool
	dbConnection       bool
	pyxisEventsLoaded  bool
}

func (p *processState) UpdateState() {
	if p.mode == TestMode {
		return
	}

	if !p.erxs {
		p.mode = SafetyMode
		return
	}

	if !p.eRxItemIdLinks {
		p.mode = SafetyMode
		return
	}

	if !p.pyxisEventsLoaded {
		p.mode = SafetyMode
		return
	}

	if !p.eRxItemIdLinks {
		p.mode = SafetyMode
		return
	}

	if !p.departmentCoverage {
		p.mode = SafetyMode
	}

	if !p.dbConnection {
		p.mode = LocalOnly
		return
	}

	p.mode = Normal
}

func (p *processState) Mode() processMode {
	return p.mode
}

func (p *processState) GetState() string {
	res := ""

	res = res + "Mode: " + strconv.Itoa(int(p.mode)) + "\n\n"
	res = res + "Pyxis Events: " + boolStatus(p.pyxisEventsLoaded) + "\n"
	res = res + "Database Connection: " + boolStatus(p.dbConnection) + "\n"
	res = res + "ERxItemIdLinks: " + boolStatus(p.eRxItemIdLinks) + "\n"
	res = res + "Department Coverage: " + boolStatus(p.departmentCoverage) + "\n"

	return res
}

func (p *processState) ERXsOkay() bool {
	return p.erxs
}

func (p *processState) ERXsLoadSuccessful() {
	p.erxs = true
	p.UpdateState()
}

func (p *processState) ERxItemIdLinksOkay() bool {
	return p.eRxItemIdLinks
}

func (p *processState) ERxItemIdLinksSuccessful() {
	p.eRxItemIdLinks = true
	p.UpdateState()
}

func (p *processState) DepartmentCoverageOkay() bool {
	return p.departmentCoverage
}

func (p *processState) DepartmentCoverageSuccessful() {
	p.departmentCoverage = true
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

func boolStatus(status bool) string {
	if status {
		return "Good"
	} else {
		return "Failed"
	}
}
