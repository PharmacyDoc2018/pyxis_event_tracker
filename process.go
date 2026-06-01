package main

import (
	"database/sql"
	"fmt"
	"os"
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

type Process struct {
	PyxisEventLogs     []PyxisEventLog
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
		p.findMissingPyxisEvents()

	} else {
		p.logger.LogInfo("Startup log checks skipped - no database connection")
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
		godotenv.Load(".env")
	}

	connString = os.Getenv("CONNSTRING")
	processLogPath = os.Getenv("PROCESSLOGPATH")
	p.pathToData = os.Getenv("DATAPATH")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("error creating connection pool: %s\n ", err.Error())
	} else {
		p.state.DbConnectionSuccessful()
	}
	p.db = db

	p.dbq = database.New(db)

	p.cacheStop = make(chan struct{})
	p.cache = cache.NewCache(cacheInterval, p.cacheStop)

	processLogger := initProcessLogger(processLogPath)
	p.logger = processLogger
	p.logger.LogInfo("Application Started")

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
		// need to halt startup and safely exit
	} else {
		p.state.ERxItemIdLinksSuccessful()
	}

	p.departmentCoverage = initDepartmentCoverage()

	return &p
}

func (p *Process) exit() {
	p.logger.LogInfo("Closing Application...")

	if p.state.PyxisEventLogsLoadedOkay() {
		p.savePyxisEventLogs()
	} else {
		p.logger.LogInfo("Pyxis event logs not being saved due to previous load error")
	}

	if p.state.ERxItemIdLinksOkay() {
		p.saveERxItemIdLinks()
	} else {
		p.logger.LogInfo("ERx - ItemId links not being saved due to previous load error")
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
	defer env.Close()

	_, err = env.WriteString(defaultEnv)
	if err != nil {
		return err
	}

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

	err = os.Mkdir("./data/log_settings/", 0755)
	if err != nil {
		return err
	}

	err = os.Mkdir("./data/pyxis_event_logs/", 0755)
	if err != nil {
		return err
	}

	return nil

}
