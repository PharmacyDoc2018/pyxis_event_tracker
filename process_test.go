package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPyxisEventLog(t *testing.T) {
	startTime, err := time.Parse("01/02/2006", "01/01/2026")
	if err != nil {
		t.Fatalf("Failed to make start time: %s", err.Error())
	}

	firstEventTime, _ := time.Parse("01/02/2006 15:04", "01/02/2026 09:48")
	secondEventTime, _ := time.Parse("01/02/2006 15:04", "01/04/2026 13:23")
	thirdEventTime, _ := time.Parse("01/02/2006 15:04", "01/06/2026 10:56")

	testPyxis := createNewPyxisEventLog("testPyxis", startTime)
	testPyxis.addEvents([]PyxisEvent{
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, One",
			UserID:                "abc1a",
			StorageSpace:          "TESTPYXIS_MAIN Drw 4.1-Pkt E4",
			ItemID:                "07571",
			MedClassCode:          "2",
			MedDisplayName:        "oxyCODONE 5 mg TABLET UD",
			TransactionType:       "Count inventory",
			TxDateTime:            firstEventTime,
			EnteredQuantity:       96.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      "480 mg",
			BegInventory:          96.0000,
			EndInventory:          96.0000,
			WitnessName:           "Testnurse, Two",
			WitnessID:             "abc2b",
			MRN:                   "",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Two",
			UserID:                "abc2b",
			StorageSpace:          "TESTPYXIS_MAIN Drw 3.1-Pkt D3",
			ItemID:                "06896",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "PROCHLORPERAZINE 10 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            secondEventTime,
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      "10 mg",
			BegInventory:          16.0000,
			EndInventory:          15.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "1234567",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Three",
			UserID:                "abc3c",
			StorageSpace:          "TESTPYXIS_MAIN Drw 2.1-Pkt A2",
			ItemID:                "05018",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "diphenhydrAMINE 50 mg (1 mL) VIAL",
			TransactionType:       "Remove",
			TxDateTime:            thirdEventTime,
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      "50 mg",
			BegInventory:          4.0000,
			EndInventory:          3.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "7654321",
		},
	})

	if len(testPyxis.Log) != 3 {
		t.Fatalf("error. AddEvents method failed to add pyxis events")
	}

	newEvents := []PyxisEvent{
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, One",
			UserID:                "abc1a",
			StorageSpace:          "TESTPYXIS_MAIN Drw 4.1-Pkt E5",
			ItemID:                "49375",
			MedClassCode:          "3",
			MedDisplayName:        "lorazepam 0.5 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            firstEventTime.Add(24 * time.Hour),
			EnteredQuantity:       1.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      "0.5 mg",
			BegInventory:          13.0000,
			EndInventory:          12.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "2143657",
		},
		{
			ItemTransactionKey:    uuid.New(),
			UserName:              "Testnurse, Two",
			UserID:                "abc2b",
			StorageSpace:          "TESTPYXIS_MAIN Drw 5.1-Pkt E6",
			ItemID:                "98367",
			MedClassCode:          "Non-Controlled",
			MedDisplayName:        "acetaminophen 325 mg TABLET UD",
			TransactionType:       "Remove",
			TxDateTime:            secondEventTime.Add(24 * time.Hour),
			EnteredQuantity:       2.0000,
			EnteredUOMDisplayCode: "Dosage Form",
			AmountReferenced:      "0.5 mg",
			BegInventory:          97.0000,
			EndInventory:          95.0000,
			WitnessName:           "None/Unknown",
			WitnessID:             "Unknown",
			MRN:                   "9876543",
		},
	}

	testPyxis.addEvents(newEvents)
	expectedOrder := []string{
		"oxyCODONE 5 mg TABLET UD",
		"lorazepam 0.5 mg TABLET UD",
		"PROCHLORPERAZINE 10 mg TABLET UD",
		"acetaminophen 325 mg TABLET UD",
		"diphenhydrAMINE 50 mg (1 mL) VIAL",
	}

	for i, event := range testPyxis.Log {
		if event.MedDisplayName != expectedOrder[i] {
			t.Errorf("error. expected %s. found %s", expectedOrder[i], event.MedDisplayName)
		}
	}

}
