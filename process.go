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

type ProcessState struct {
	PyxisEventLogs []PyxisEventLog
	pathToData     string
	logger         processLogger
	erxItemIdLinks *ERxItemIdLinks
	db             *sql.DB
	dbq            *database.Queries
	cliConfig      *cli.Config
	cache          *cache.Cache
	cacheStop      chan struct{}
	dbConnection   bool
}

func (p *ProcessState) startupLogsCheck() {
	if p.dbConnection {
		p.logger.LogInfo("Initiating startup logs check")
		p.findMissingPyxisEvents()

	} else {
		p.logger.LogInfo("Startup log checks skipped - no database connection")
	}
}

func initProcess() *ProcessState {
	p := ProcessState{}
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
		// need to halt startup and safely exit
	}

	p.erxItemIdLinks.Map = make(map[string]ERxItemIdLink)
	err = p.loadERxItemIdLinks()
	if err != nil {
		fmt.Println(err.Error())
		// need to halt startup and safely exit
	}

	return &p
}

func (p *ProcessState) exit() {
	p.logger.LogInfo("Closing Application...")
	p.savePyxisEventLogs()
	p.cliConfig.Rl.Close()
	close(p.cacheStop)
	time.Sleep(500 * time.Millisecond)
	p.logger.LogInfo("Application Closed")
	p.logger.Close()
	os.Exit(0)
}

const defaultEnv = `CONNSTRING=""
PROCESSLOGPATH="./logs/process_log.txt"
DATAPATH="./data/"
`

func (p *ProcessState) initialLaunchSetup() error {
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
