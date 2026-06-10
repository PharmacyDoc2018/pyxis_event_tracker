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

	p.startupLogsCheck()
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
