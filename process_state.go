package main

type processMode int

const (
	Normal = iota
	LocalOnly
	SafetyMode
)

type processState struct {
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

func initProcessState() *processState {
	return &processState{
		mode: SafetyMode,
	}
}
