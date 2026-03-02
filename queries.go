package main

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

type MedicationName struct {
	MedDisplayName string
}

const listControlTwoMedsByDevice = `
SELECT DISTINCT i.MedDisplayName
FROM PYX.fctItemTransaction t
INNER JOIN PYX.dimDispensingDevice d
ON t.UserAtDispensingDeviceKey = d.DispensingDeviceKey
INNER JOIN PYX.dimItem i
ON t.ItemKey = i.ItemKey
WHERE d.DispensingDeviceName = @device
AND i.MedClassCode = '2'
`

func (q *Queries) ListControlTwoMedsByDevice(ctx context.Context, device string) ([]MedicationName, error) {
	rows, err := q.db.QueryContext(ctx, listControlTwoMedsByDevice, sql.Named("device", device))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MedicationName
	for rows.Next() {
		var i MedicationName
		if err := rows.Scan(
			&i.MedDisplayName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
