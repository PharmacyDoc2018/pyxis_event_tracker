package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cache"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/joho/godotenv"
)

const cacheInterval = 60 * time.Minute
const minPyxisEventRecheckInterval = 24 * time.Hour
const pyxisEventLogsFolder = "pyxis_event_logs"
const pyxisEventLogSettingsFolder = "log_settings"
const controlEventLogsFolder = "control_logs"

type Process struct {
	PyxisEventLogs     []*PyxisEventLog
	pathToData         string
	pathToSettings     string
	pathToOut          string
	logger             processLogger
	testMarActionRes   []database.MarActionResponse
	state              *processState
	settings           *Settings
	erxs               *ERxDict
	itemIDs            *ItemIdDict
	erxItemIdLinks     *ERxItemIdLinks
	departmentCoverage *DepartmentCoverage
	db                 *sql.DB
	dbq                *database.Queries
	cliConfig          *cli.Config
	cache              *cache.Cache
	cacheStop          chan struct{}
	dbConnection       bool
}

func (p *Process) startupLogsCheck() {
	if p.state.DbConnectionOkay() {
		p.logger.LogInfo("Initiating startup logs check")

		//-- Check for nil PyxisEventLog pointers in ControlEventLog.
		p.logger.LogInfo("Checking ControlEventLog connections")
		for i, log := range p.PyxisEventLogs {
			if log.ControlEventLog.pyxisEventLog == nil {
				log.ControlEventLog.pyxisEventLog = p.PyxisEventLogs[i]
				p.logger.LogInfo(fmt.Sprintf("PyxisEventLog pointer created for %s ControlEventLog", log.PyxisName))
			}
		}
		p.findMissingPyxisEvents()

	} else {
		p.logger.LogInfo("Startup log checks skipped - no database connection")
	}
}

func (p *Process) cleanUpPyxisEventLogs() {
	for _, pyxisEventLog := range p.PyxisEventLogs {
		logger := pyxisEventLog.CleanUp()
		p.logger.Log(logger)
	}
}

func (p *Process) findMissingPyxisEvents() {
	for i := range p.PyxisEventLogs {
		startTime := time.Time{}
		if p.PyxisEventLogs[i].LastEventDateTime.IsZero() {
			startTime = p.PyxisEventLogs[i].StartDateTime
		} else {
			startTime = p.PyxisEventLogs[i].LastEventDateTime
		}

		endTime := timeToday() //--time today at midnight for cache if duplicate call

		if endTime.Sub(startTime) < minPyxisEventRecheckInterval {
			p.logger.LogInfo(fmt.Sprintf("Last Pyxis event for %s on %s, less than 24 hours ago. Finding missing Pyxis events skipped",
				p.PyxisEventLogs[i].PyxisName,
				startTime.Format("2006-01-02 1504")))
			continue
		}

		params := database.GetPyxisEventsForDeviceByDateRangeParams{
			Device: p.PyxisEventLogs[i].PyxisName,
			Start:  startTime,
			End:    endTime,
		}

		events, err := getPyxisEvents(p, params) //-- logging handled in function
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		p.logger.Log(p.PyxisEventLogs[i].ParseEventsAndAdd(events))
	}
}

func (p *Process) createNewPyxisEventLog(pyxisName string, startDateTime time.Time) error {
	for _, pyxisLog := range p.PyxisEventLogs {
		if pyxisName == pyxisLog.PyxisName {
			err := fmt.Errorf("error. %s already exists", pyxisName)
			p.logger.LogError(fmt.Sprintf("Error. Failed to create new Pyxis event log: %s", err.Error()))
			return err
		}
	}

	newPyxisLog := &PyxisEventLog{
		Log:           []PyxisEvent{},
		StartDateTime: startDateTime,
		PyxisName:     pyxisName,
	}
	newPyxisLog.ControlEventLog = &ControlEventLog{
		pyxisEventLog: newPyxisLog,
	}

	p.PyxisEventLogs = append(p.PyxisEventLogs, newPyxisLog)
	p.logger.LogInfo(fmt.Sprintf("New Pyxis event log: %s added. Logging events starting on or after %s.",
		pyxisName,
		startDateTime.Format("2006-01-02 1504")))

	p.state.PyxisEventLogLoaded(pyxisName)

	return nil
}

func (p *Process) matchControlEventActions() {
	for i := range p.PyxisEventLogs {
		p.logger.LogInfo(fmt.Sprintf("Starting control event matching for %s", p.PyxisEventLogs[i].PyxisName))

		//-- If no unmatched events, skip to next Pyxis
		if len(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents) == 0 {
			p.logger.LogInfo(fmt.Sprintf("No unmatched control events for %s", p.PyxisEventLogs[i].PyxisName))
			continue
		}

		coveredDeptIDs, logErr := p.departmentCoverage.GetCoveredDepartments(p.PyxisEventLogs[i].PyxisName)
		if logErr != nil {
			p.logger.LogError(logErr.logMessage + " Unable to match control events")
			continue
		}
		p.logger.LogInfo(fmt.Sprintf("Found %d departments covered by %s",
			len(coveredDeptIDs),
			p.PyxisEventLogs[i].PyxisName))

		//-- Sort unmatched control events by datetime
		p.logger.LogInfo(fmt.Sprintf("Sorting %d unmatched control events for %s",
			len(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents),
			p.PyxisEventLogs[i].PyxisName))

		p.PyxisEventLogs[i].ControlEventLog.SortUnmatchedEvents()

		//-- Get variables needed from sorted unmatched events for MAR action query
		firstDay := timeStartDay(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents[0].TxDateTime)
		lastDay := timeEndDay(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents[len(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents)-1].TxDateTime)

		mrns := []string{}
		mrnMap := map[string]struct{}{}

		itemIDsMap := map[string]struct{}{}

		medIDs := []string{}

		for _, event := range p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents {
			mrnMap[event.MRN] = struct{}{}
			itemIDsMap[event.ItemID] = struct{}{}
		}

		for key := range mrnMap {
			mrns = append(mrns, key)
		}
		mrnMap = nil

		for key := range itemIDsMap {
			medIDs = append(medIDs, p.erxItemIdLinks.GetMedIds(key)...)
		}
		itemIDsMap = nil

		deptIDs := []string{}
		for _, dept := range coveredDeptIDs {
			deptIDs = append(deptIDs, dept.ID)
		}

		unmatchedEvents := p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents
		p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents = []PyxisEvent{}

		params := database.GetMarAdminActionsByPatientsDaysMedIDsParams{
			DateStart: firstDay,
			DateEnd:   lastDay,
			DeptIDs:   deptIDs,
			Mrns:      mrns,
			MedIDs:    medIDs,
		}

		p.logger.LogInfo(fmt.Sprintf("Getting MAR actions from %s to %s for mrns %s and medIDs %s",
			firstDay.Format("2006-01-02 1504"),
			lastDay.Format("2006-01-02 1504"),
			strings.Join(mrns, ", "),
			strings.Join(medIDs, ", ")))

		MarActionResponses, err := getMarActions(p, params)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Unable retrieve MAR actions for %s Pyxis from %s to %s - control event matching skipped",
				p.PyxisEventLogs[i].PyxisName,
				firstDay.Format("2006-01-02"),
				lastDay.Format("2006-01-02")))

			continue
		}

		marActions := p.parseMarActions(MarActionResponses)

		eventMap := map[time.Time]map[string]map[string][]PyxisEvent{}
		for _, unmatchedEvent := range unmatchedEvents {
			startDayTime := timeStartDay(unmatchedEvent.TxDateTime)
			mrn := unmatchedEvent.MRN
			itemID := unmatchedEvent.ItemID

			if _, okay := eventMap[startDayTime]; !okay {
				eventMap[startDayTime] = map[string]map[string][]PyxisEvent{}
			}

			if _, okay := eventMap[startDayTime][mrn]; !okay {
				eventMap[startDayTime][mrn] = map[string][]PyxisEvent{}
			}

			if _, okay := eventMap[startDayTime][mrn][itemID]; !okay {
				eventMap[startDayTime][mrn][itemID] = []PyxisEvent{}
			}

			eventMap[startDayTime][mrn][itemID] = append(eventMap[startDayTime][mrn][itemID], unmatchedEvent)
		}

		actionMap := map[time.Time]map[string]map[string][]MarAction{}
		for _, marAction := range marActions {
			startDayTime := timeStartDay(marAction.SavedTime)
			mrn := marAction.MRN
			itemID, logErr := p.erxItemIdLinks.GetItemId(marAction.MedicationID)
			if logErr != nil {
				p.logger.LogError(fmt.Sprintf("Error. no itemID linked to medID %s. No MAR actions will be matched for mrn %s on %s for that medID",
					marAction.MedicationID,
					marAction.MRN,
					startDayTime.Format("2006-01-02")))
				continue
			}

			if _, okay := actionMap[startDayTime]; !okay {
				actionMap[startDayTime] = map[string]map[string][]MarAction{}
			}

			if _, okay := actionMap[startDayTime][mrn]; !okay {
				actionMap[startDayTime][mrn] = map[string][]MarAction{}
			}

			if _, okay := actionMap[startDayTime][mrn][itemID]; !okay {
				actionMap[startDayTime][mrn][itemID] = []MarAction{}
			}

			actionMap[startDayTime][mrn][itemID] = append(actionMap[startDayTime][mrn][itemID], marAction)
		}

		for startDayTime := range eventMap {
			for mrn := range eventMap[startDayTime] {
				for itemId := range eventMap[startDayTime][mrn] {
					p.logger.LogInfo(fmt.Sprintf("Matching %d Pyxis events and %d MAR actions for mrn %s on %s with itemID %s",
						len(eventMap[startDayTime][mrn][itemId]),
						len(actionMap[startDayTime][mrn][itemId]),
						mrn,
						startDayTime.Format("2006-01-02"),
						itemId))

					unmatchedEvents := p.PyxisEventLogs[i].ControlEventLog.MatchEvents(eventMap[startDayTime][mrn][itemId], actionMap[startDayTime][mrn][itemId], startDayTime, mrn, itemId)
					if len(unmatchedEvents) > 0 {
						p.logger.LogInfo(fmt.Sprintf("Matching complete with %d unmatched events", len(unmatchedEvents)))
						p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents = append(p.PyxisEventLogs[i].ControlEventLog.UnmatchedEvents, unmatchedEvents...)
					} else {
						p.logger.LogInfo("Match complete with no unmatched events")
					}
				}
			}
		}
	}
}

func initProcess() *Process {
	p := Process{}
	p.state = initProcessState()

	connString := ""
	processLogPath := ""

	err := godotenv.Load(".env")
	if err != nil {
		err := p.initialLaunchSetup()
		if err != nil {
			fmt.Printf("ERROR. first time initilization process not complete: %s\n", err.Error())
		}
	}

	connString = os.Getenv("CONNSTRING")
	processLogPath = os.Getenv("PROCESSLOGPATH")
	p.pathToData = os.Getenv("DATAPATH")
	p.pathToSettings = os.Getenv("SETTINGSPATH")
	p.pathToOut = os.Getenv("OUTPATH")

	processLogger := initProcessLogger(processLogPath)
	p.logger = processLogger
	p.logger.LogInfo("Application Started")

	err = p.loadSettings()
	if err != nil {
		fmt.Println(err.Error())
	}
	p.logger.printToIO = p.settings.PrintLogsToCliIO

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("error creating connection pool: %s\n ", err.Error())
	}

	p.db = db
	p.dbq = database.New(db)

	p.cacheStop = make(chan struct{})
	p.cache = cache.NewCache(cacheInterval, p.cacheStop)

	fmt.Println("Attempting to connect to database...")
	err = p.db.Ping()
	if err != nil {
		fmt.Println("connection failed!")
		fmt.Println("warning: no database connection")
		p.state.DbConnectionFail()
		p.logger.LogError("Connection to database not successful")
	} else {
		fmt.Println("connection successful")
		p.state.DbConnectionSuccessful()
		p.logger.LogInfo("Connection to database successful")
	}

	err = p.loadPyxisEventLogs()
	if err != nil {
		fmt.Printf("error loading Pyxis event logs: %s\n", err.Error())
	} else {
		p.state.PyxisEventLogsLoadSuccessful()
	}

	p.erxItemIdLinks = initERxItemIdLink()
	err = p.loadERxItemIdLinks()
	if err != nil {
		fmt.Println(err.Error())
	}

	p.departmentCoverage = initDepartmentCoverage()
	err = p.loadDepartmentCoverage()
	if err != nil {
		fmt.Println(err.Error())
	}

	p.erxs = initERXs()
	err = p.erxs.Load(&p)
	if err != nil {
		fmt.Println(err.Error())
	}

	p.itemIDs = initItemIDs()
	err, lr := p.itemIDs.Load(p.pathToData)
	p.logger.Log(lr)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		p.state.ItemIDsLoadSuccessful()
	}

	//-- Check for new Pyxis events
	p.startupLogsCheck()

	if !p.state.DbConnectionOkay() {
		p.logger.LogInfo("New control event check and matched skipped - no database connection")
	} else {
		//-- Check for new control events -> add them to their ControlEventLog.UnmatchedEvents
		for i := range p.PyxisEventLogs {
			p.logger.Log(p.PyxisEventLogs[i].checkForNewControlEvents())
		}

		//-- Attempt to match events in ControlEventLog.UnmatchedEvents
		p.matchControlEventActions()

		//-- Validate control event trails
		for i := range p.PyxisEventLogs {
			p.logger.LogInfo(fmt.Sprintf("Validating control event trails for %s Pyxis", p.PyxisEventLogs[i].PyxisName))
			p.PyxisEventLogs[i].ControlEventLog.ValidateTrails()
		}

	}

	//-- unload Pyxis event logs from memory
	p.saveAndUnloadPyxisEventLogs()

	return &p
}

func (p *Process) exit() {
	p.logger.LogInfo("Closing Application...")

	if p.state.PyxisEventLogsLoadedOkay() {
		err := p.saveAndUnloadPyxisEventLogs()
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("Pyxis event logs not being saved due to previous load error")
	}

	if p.state.ERxItemIdLinksOkay() {
		err := p.saveERxItemIdLinks()
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ERx - ItemId links not being saved due to previous load error")
	}

	if p.state.DepartmentCoverageOkay() {
		err := p.saveDepartmentCoverage()
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("Department - Coverage links not being saved due to previous load error")
	}

	if p.state.ERXsOkay() {
		err := p.erxs.Save(p)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error while saving: %s", err.Error()))
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ERXs not being saved due to previous load error")
	}

	if p.state.ItemIDsOkay() {
		err, lr := p.itemIDs.Save(p.pathToData)
		p.logger.Log(lr)
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		p.logger.LogInfo("ItemIDs not being saved due to previous load error")
	}

	p.cliConfig.Rl.Close()
	close(p.cacheStop)
	time.Sleep(500 * time.Millisecond)
	p.logger.LogInfo("Application Closed")
	p.logger.EndSpace()
	p.logger.Close()
	os.Exit(0)
}

const defaultEnv = `CONNSTRING=""
PROCESSLOGPATH="./logs/process_log.txt"
DATAPATH="./data/"
SETTINGSPATH="./settings/"
OUTPATH="./out/"
`

func (p *Process) initialLaunchSetup() error {
	env, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = env.WriteString(defaultEnv)
	if err != nil {
		return err
	}

	env.Close()
	godotenv.Load(".env")
	p.pathToData = os.Getenv("DATAPATH")
	p.pathToSettings = os.Getenv("SETTINGSPATH")
	p.pathToOut = os.Getenv("OUTPATH")

	err = os.Mkdir("./logs/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir("./data/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(p.pathToSettings, 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(p.pathToOut, 0755)
	if err != nil {
		return err
	}

	f, err := os.OpenFile("./data/ERxItemIdLinks.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString("{}")
	if err != nil {
		return err
	}
	f.Close()

	f, err = os.OpenFile(filepath.Join(p.pathToData, departmentCoverageFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString("{}")
	if err != nil {
		return err
	}
	f.Close()

	err = os.Mkdir("./data/log_settings/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir("./data/pyxis_event_logs/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(p.pathToData, controlEventLogsFolder), 0755)
	if err != nil {
		return err
	}

	return nil

}
