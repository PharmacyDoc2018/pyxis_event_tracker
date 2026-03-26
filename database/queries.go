package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

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

const getPyxisEventsForDeviceByDateRange = `
SELECT
	t.ItemTransactionKey,
	u.FullName AS UserName,
	u.UserID,
	space.StorageSpaceShortName AS StorageSpace,
	i.ItemID,
	i.MedClassCode,
	i.MedDisplayName,
	tc.TransactionType,
	t.Tx_Date AS TxDate,
	time.string_representation_24 AS TxTime,
	t.EnteredQuantity,
	t.EnteredUOMDisplayCode,
	CAST(
		CASE
			WHEN t.EnteredUOMDisplayCode IS NULL
				THEN NULL
			WHEN t.EnteredUOMDisplayCode = 'Dosage Form'
				THEN t.EnteredQuantity * i.StrengthAmount
			WHEN t.EnteredUOMDisplayCode = i.VolumeUnit_DisplayCode
				THEN (t.EnteredQuantity * i.StrengthAmount) / i.TotalVolumeAmount
			ELSE
				t.EnteredQuantity
		END
		AS numeric(18,4)
	) AS AmountReferenced,
	i.StrengthUnit_DisplayCode AS AmountReferencedUnits,
	t.BegInventory,
	t.EndInventory,
	w.FullName AS WitnessName,
	w.UserID AS WitnessID,
	p.PatientID AS MRN
FROM PYX.fctItemTransaction t
INNER JOIN PYX.dimDispensingDevice d
	ON t.UserAtDispensingDeviceKey = d.DispensingDeviceKey
INNER JOIN PYX.dimItem i
	ON t.ItemKey = i.ItemKey
INNER JOIN dbo.dimTime time
	ON t.Tx_TimeKey = time.DimTime_key
LEFT JOIN PYX.fctItemEncounters e
	ON t.EncounterKey = e.EncounterKey
INNER JOIN PYX.dimUserAccount u
	ON t.UserAccountKey = u.UserAccountKey
INNER JOIN PYX.dimUserAccount w
	ON t.WitnessAccountKey = w.UserAccountKey
INNER JOIN PYX.dimTransactionCharacteristics tc
	ON t.sk_dim_TxChar = tc.sk_dim_TxChar
LEFT JOIN PYX.fctPatient p
	ON e.PatientKey = p.PatientKey
LEFT JOIN PYX.dimStorageSpace space
	ON t.StorageSpaceKey = space.StorageSpaceKey
WHERE d.DispensingDeviceName = @device
AND t.TransactionLocalDateTime >= @start
AND t.TransactionLocalDateTime < @end
AND NOT (
	tc.TransactionType = 'Remove'
	AND
	t.BegInventory IS NULL
)
ORDER BY
	t.Tx_Date,
	time.string_representation_24,
	e.LastItemTransactionLocalDateTime;
`

type PyxisEventResponse struct {
	ItemTransactionKey    uuid.UUID
	UserName              sql.NullString
	UserID                sql.NullString
	StorageSpace          sql.NullString
	ItemID                sql.NullString
	MedClassCode          sql.NullString
	MedDisplayName        sql.NullString
	TransactionType       sql.NullString
	TxDate                sql.NullTime
	TxTime                sql.NullString
	EnteredQuantity       sql.NullFloat64
	EnteredUOMDisplayCode sql.NullString
	AmountReferenced      sql.NullFloat64
	AmountReferencedUnits sql.NullString
	BegInventory          sql.NullFloat64
	EndInventory          sql.NullFloat64
	WitnessName           sql.NullString
	WitnessID             sql.NullString
	MRN                   sql.NullString
}

type GetPyxisEventsForDeviceByDateRangeParams struct {
	Device string
	Start  string
	End    string
}

func (q *Queries) GetPyxisEventsForDeviceByDateRange(ctx context.Context, arg GetPyxisEventsForDeviceByDateRangeParams) ([]PyxisEventResponse, error) {
	startDate, err := parseDate(arg.Start)
	if err != nil {
		return nil, fmt.Errorf("error. unable to parse start date: %s", err.Error())
	}

	endDate, err := parseDate(arg.End)
	if err != nil {
		return nil, fmt.Errorf("error. unable to parse end date: %s", err.Error())
	}

	rows, err := q.db.QueryContext(ctx, getPyxisEventsForDeviceByDateRange, sql.Named("device", arg.Device), sql.Named("start", startDate), sql.Named("end", endDate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ct := 0
	var items []PyxisEventResponse
	for rows.Next() {
		ct++
		var i PyxisEventResponse
		if err := rows.Scan(
			&i.ItemTransactionKey,
			&i.UserName,
			&i.UserID,
			&i.StorageSpace,
			&i.ItemID,
			&i.MedClassCode,
			&i.MedDisplayName,
			&i.TransactionType,
			&i.TxDate,
			&i.TxTime,
			&i.EnteredQuantity,
			&i.EnteredUOMDisplayCode,
			&i.AmountReferenced,
			&i.AmountReferencedUnits,
			&i.BegInventory,
			&i.EndInventory,
			&i.WitnessName,
			&i.WitnessID,
			&i.MRN,
		); err != nil {
			return nil, err
		}
		fmt.Printf("%d items found from query\n", ct)
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
