package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cache"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// -- github.com/gocarina/gocsv
const cacheInterval = 60 * time.Minute

type ProcessState struct {
	PyxisUnitsLogs []PyxisEventLog
	logger         processLogger
	db             *sql.DB
	dbq            *database.Queries
	cliConfig      *cli.Config
	cache          *cache.Cache
	cacheStop      chan struct{}
	dbConnection   bool
}

func (p *ProcessState) findMissingEvents() {
	for i := range p.PyxisUnitsLogs {
		params := database.GetPyxisEventsForDeviceByDateRangeParams{
			Device: p.PyxisUnitsLogs[i].PyxisName,
			Start:  p.PyxisUnitsLogs[i].LastEventDateTime,
			End:    time.Now(),
		}

		events, err := getPyxisEvents(p, params)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		p.PyxisUnitsLogs[i].parseEventsAndAdd(events)
	}
}

type PyxisEvent struct {
	ItemTransactionKey    uuid.UUID
	UserName              string
	UserID                string
	StorageSpace          string
	ItemID                string
	MedClassCode          string
	MedDisplayName        string
	TransactionType       string
	TxDateTime            time.Time
	EnteredQuantity       float64
	EnteredUOMDisplayCode string
	AmountReferenced      float64
	AmountReferencedUnits string
	BegInventory          float64
	EndInventory          float64
	WitnessName           string
	WitnessID             string
	MRN                   string
}

type PyxisEventLog struct {
	Log               []PyxisEvent
	StartDateTime     time.Time
	LastEventDateTime time.Time
	PyxisName         string
}

func (p *PyxisEventLog) cleanUp() {
	//-- resort the events
	fmt.Printf("sorting %s event log\n...", p.PyxisName)
	sort.Slice(p.Log, func(i, j int) bool {
		return p.Log[i].TxDateTime.Before(p.Log[j].TxDateTime)
	})
	fmt.Println("sorting complete!")

	//-- check for duplicates
	fmt.Println("checking for duplicates")
	newLog := []PyxisEvent{}
	numDups := 0
	newLog = append(newLog, p.Log[0])
	for i := 1; i < len(p.Log); i++ {
		if p.Log[i] == p.Log[i-1] {
			numDups++
			continue
		} else {
			newLog = append(newLog, p.Log[i])
		}
	}
	p.Log = newLog
	switch numDups {
	case 0:
		fmt.Println("check complete! no duplicates found")

	case 1:
		fmt.Println("check complete! 1 duplicate removed")

	default:
		fmt.Printf("check complete! %d duplicates removed\n", numDups)
	}

	//-- update date range
	p.LastEventDateTime = p.Log[len(p.Log)-1].TxDateTime
}

func (p *PyxisEventLog) addEvents(events []PyxisEvent) {
	p.Log = append(p.Log, events...)
	p.cleanUp()
}

func (p *PyxisEventLog) lastEventDateString() string {
	if p.LastEventDateTime.IsZero() {
		return ""
	}

	return p.LastEventDateTime.Format("2006-01-02 15:04")
}

func (p *PyxisEventLog) parseEventsAndAdd(events []database.PyxisEventResponse) {
	parsedEvents := []PyxisEvent{}

	for _, event := range events {
		pyxisEvent := PyxisEvent{}
		pyxisEvent.ItemTransactionKey = event.ItemTransactionKey

		if event.UserName.Valid {
			pyxisEvent.UserName = event.UserName.String
		} else {
			pyxisEvent.UserName = ""
		}

		if event.UserID.Valid {
			pyxisEvent.UserID = event.UserID.String
		} else {
			pyxisEvent.UserID = ""
		}

		if event.StorageSpace.Valid {
			pyxisEvent.StorageSpace = event.StorageSpace.String
		} else {
			pyxisEvent.StorageSpace = ""
		}

		if event.ItemID.Valid {
			pyxisEvent.ItemID = event.ItemID.String
		} else {
			pyxisEvent.ItemID = ""
		}

		if event.MedClassCode.Valid {
			pyxisEvent.MedClassCode = event.MedClassCode.String
		} else {
			pyxisEvent.MedClassCode = ""
		}

		if event.MedDisplayName.Valid {
			pyxisEvent.MedDisplayName = event.MedDisplayName.String
		} else {
			pyxisEvent.MedDisplayName = ""
		}

		if event.TransactionType.Valid {
			pyxisEvent.TransactionType = event.TransactionType.String
		} else {
			pyxisEvent.TransactionType = ""
		}

		if event.TxDateTime.Valid {
			pyxisEvent.TxDateTime = event.TxDateTime.Time
		} else {
			pyxisEvent.TxDateTime = time.Time{}
		}

		if event.EnteredQuantity.Valid {
			pyxisEvent.EnteredQuantity = event.EnteredQuantity.Float64
		} else {
			pyxisEvent.EnteredQuantity = 0.0000
		}

		if event.EnteredUOMDisplayCode.Valid {
			pyxisEvent.EnteredUOMDisplayCode = event.EnteredUOMDisplayCode.String
		} else {
			pyxisEvent.EnteredUOMDisplayCode = ""
		}

		if event.AmountReferenced.Valid {
			pyxisEvent.AmountReferenced = event.AmountReferenced.Float64
		} else {
			pyxisEvent.AmountReferenced = 0.0000
		}

		if event.AmountReferencedUnits.Valid {
			pyxisEvent.AmountReferencedUnits = event.AmountReferencedUnits.String
		} else {
			pyxisEvent.AmountReferencedUnits = ""
		}

		if event.BegInventory.Valid {
			pyxisEvent.BegInventory = event.BegInventory.Float64
		} else {
			pyxisEvent.BegInventory = 0.0000
		}

		if event.EndInventory.Valid {
			pyxisEvent.EndInventory = event.EndInventory.Float64
		} else {
			pyxisEvent.EndInventory = 0.0000
		}

		if event.WitnessName.Valid {
			pyxisEvent.WitnessName = event.WitnessName.String
		} else {
			pyxisEvent.WitnessName = ""
		}

		if event.WitnessID.Valid {
			pyxisEvent.WitnessID = event.WitnessID.String
		} else {
			pyxisEvent.WitnessID = ""
		}

		if event.MRN.Valid {
			pyxisEvent.MRN = event.MRN.String
		} else {
			pyxisEvent.MRN = ""
		}

		parsedEvents = append(parsedEvents, pyxisEvent)
	}

	p.addEvents(parsedEvents)

}

func createNewPyxisEventLog(pyxisName string, startDateTime time.Time) PyxisEventLog {
	return PyxisEventLog{
		Log:           []PyxisEvent{},
		StartDateTime: startDateTime,
		PyxisName:     pyxisName,
	}
}

func initProcess() *ProcessState {
	p := ProcessState{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")
	processLogPath := os.Getenv("PROCESSLOGPATH")

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

	return &p
}

func (p *ProcessState) exit() {
	p.logger.LogInfo("Application Closed")
	p.cliConfig.Rl.Close()
	p.logger.Close()
	close(p.cacheStop)
	time.Sleep(500 * time.Millisecond)
	os.Exit(0)
}
