package main

import (
	"database/sql"
	"fmt"
)

type EventActionRef struct {
	Type  EventType
	Index int
}

type EventActionRefTrail struct {
	Trail []EventActionRef
}

type LocalDB struct {
	db *sql.DB
}

func (l *LocalDB) GetPyxisInfo() ([]*PyxisEventLog, error) {
	type pyxisTableQueryResponse struct {
		id                      int
		name                    string
		startDateTimeString     string
		lastEventDateTimeString string
	}

	responses := []pyxisTableQueryResponse{}

	rows, err := l.db.Query(`
		SELECT * FROM pyxis
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var res pyxisTableQueryResponse
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}
		responses = append(responses, res)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	logs := []*PyxisEventLog{}
	for _, res := range responses {
		startDateTime, err := parseDate(res.startDateTimeString)
		if err != nil {
			return nil, err
		}

		lastEventDateTime, err := parseDate(res.lastEventDateTimeString)
		if err != nil {
			return nil, err
		}

		logs = append(logs, &PyxisEventLog{
			StartDateTime:     startDateTime,
			LastEventDateTime: lastEventDateTime,
			PyxisName:         res.name,
		})
	}

	return logs, nil

}

func (p *Process) InitLocalDB() error {
	db, err := sql.Open("sqlite", "app.db")
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Fatal. Unable to open local database: %s", err.Error()))
		return err
	}

	p.localdb = &LocalDB{
		db: db,
	}

	//-- Pyxis list table
	_, err = p.localdb.db.Exec(`
			CREATE TABLE IF NOT EXISTS pyxis (
				id INTEGER PRIMARY KEY AUTOINCREMENT
				name TEXT NOT NULL UNIQUE
				start_date_time TEXT NOT NULL
				last_event_date_time TEXT NOT NULL
			)
		`)
	if err != nil {
		p.logger.LogError("Failed to create pyxis table in local database")
		return err
	}

	//-- Control event table
	_, err = p.localdb.db.Exec(`
			CREATE TABLE IF NOT EXISTS control_event_ids (
				item_transaction_key TEXT PRIMARY KEY 
				user_id TEXT
				user_name TEXT
				storage_space TEXT
				item_id TEXT
				med_class_code TEXT
				med_display_name TEXT
				transaction_type TEXT
				tx_date_time TEXT
				entered_quantity REAL
				entered_uom_display_code TEXT
				amount_referenced REAL
				amount_referenced_units TEXT
				beg_inventory REAL
				end_inventory REAL
				witness_name TEXT
				witness_id TEXT
				mrn TEXT
				pyxis_id INTEGER
				matched INTEGER
			)
		`)
	if err != nil {
		p.logger.LogError("Failed to create control_event_ids table in local database")
		return err
	}

	//-- Stored MAR actions table
	_, err = p.localdb.db.Exec(`
			CREATE TABLE IF NOT EXISTS control_mar_actions (
				id INTEGER PRIMARY KEY AUTOINCREMENT
				saved_time TEXT
				order_number TEXT
				mar_action TEXT
				display_name TEXT
				medication_id TEXT
				user_id TEXT
				user_name TEXT
				dose_unit_description TEXT
				dose REAL
				mrn TEXT
				pt_name TEXT
			)
		`)
	if err != nil {
		p.logger.LogError("Failed to create control_mar_actions table in local database")
		return err
	}

	//-- Control event trails table. Data is json marshalled EventActionRefTrail
	_, err = p.localdb.db.Exec(`
		CREATE TABLE IF NOT EXISTS control_event_trails (
			id INTEGER PRIMARY KEY AUTOINCREMENT
			data BLOB NOT NULL
		)
	`)
	if err != nil {
		p.logger.LogError("Failed to create control_event_trails table in local database")
		return err
	}

	return nil
}
