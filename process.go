package main

import (
	"database/sql"
	"fmt"
	"os"
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
	PyxisUnits   []string
	db           *sql.DB
	dbq          *database.Queries
	cliConfig    *cli.Config
	cache        *cache.Cache
	cacheStop    chan struct{}
	dbConnection bool
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
	StrengthRemoved       string    //--needs conversion from float? + string
	EnteredQuantity       string    //-- needs conversion from float
	EnteredUOMDisplayCode string
	BegInventory          int
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
