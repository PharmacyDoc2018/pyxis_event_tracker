package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	db             *sql.DB
	dbq            *database.Queries
	cliConfig      *cli.Config
	cache          *cache.Cache
	cacheStop      chan struct{}
	dbConnection   bool
}

type ValType int

const (
	Null ValType = iota
	Int
	Float
)

type PyxisQuantity struct {
	Type  ValType
	Int   int64
	Float float64
}

func (pq PyxisQuantity) PrintVal() string {
	switch pq.Type {
	case Null:
		return "null"
	case Int:
		return strconv.Itoa(int(pq.Int))
	case Float:
		return strconv.FormatFloat(pq.Float, 'f', -1, 64)
	}

	return ""
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
	TxDateTime            time.Time //-- needs conversion from date + string
	EnteredQuantity       float64
	EnteredUOMDisplayCode string
	AmountReferenced      string //--needs conversion from float? + string
	BegInventory          float64
	EndInventory          float64
	WitnessName           string
	WitnessID             string
	MRN                   string
}

type PyxisEventLog struct {
	Log            []PyxisEvent
	StartDate      time.Time
	FirstEventDate time.Time
	LastEventDate  time.Time
	PyxisName      string
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
	p.FirstEventDate = p.Log[0].TxDateTime
	p.LastEventDate = p.Log[len(p.Log)-1].TxDateTime
}

func (p *PyxisEventLog) AddEvents(events []PyxisEvent) {
	p.Log = append(p.Log, events...)
	p.cleanUp()
}

func createNewPyxisEventLog(pyxisName string, startDate time.Time) *PyxisEventLog {
	return &PyxisEventLog{
		Log:       []PyxisEvent{},
		StartDate: startDate,
		PyxisName: pyxisName,
	}
}

func initProcess() *ProcessState {
	p := ProcessState{}

	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	p.db = db

	p.dbq = database.New(db)

	p.cacheStop = make(chan struct{})
	p.cache = cache.NewCache(cacheInterval, p.cacheStop)

	return &p
}

func (p *ProcessState) exit() {
	p.cliConfig.Rl.Close()
	close(p.cacheStop)
	time.Sleep(500 * time.Millisecond)
	os.Exit(0)
}
