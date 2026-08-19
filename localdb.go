package main

import (
	"database/sql"
	"fmt"
)

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

	var exists bool
	err = p.localdb.db.QueryRow(`
		SELECT EXISTS (
   	 		SELECT 1 FROM sqlite_master WHERE type='table' AND name='pyxis'
		);
	`).Scan(&exists)
	if err != nil {
		p.logger.LogError("Failed to check if pyxis table exists")
		return err
	}

	if !exists {
		//-- init and convert data
		_, err = p.localdb.db.Exec(`
			CREATE TABLE pyxis (
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

		_, err = p.localdb.db.Exec(`
			CREATE TABLE control_event_ids (
				event_id TEXT PRIMARY KEY
				pyxis_id INTEGER
			)
		`)
		if err != nil {
			p.logger.LogError("Failed to create control_event_ids table in local database")
			return err
		}

		_, err = p.localdb.db.Exec(`
			CREATE TABLE control_mar_actions (
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

		//-- Need control_event_trails table... no idea how to implement that with the varying event actions
	}

	return nil
}
