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
				event_id TEXT PRIMARY KEY
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
