package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
	logger             processLogger
	state              *processState
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

	processLogger := initProcessLogger(processLogPath)
	p.logger = processLogger
	p.logger.LogInfo("Application Started")

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

	//-- Check for new Pyxis events
	p.startupLogsCheck()

	//-- Check for new control events
	for i := range p.PyxisEventLogs {
		p.logger.Log(p.PyxisEventLogs[i].checkForNewControlEvents())
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

	err = os.Mkdir("./logs/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir("./data/", 0755)
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
