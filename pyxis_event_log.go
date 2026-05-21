package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
)

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

func (p *ProcessState) cleanUpPyxisEventLog(index int) error {
	if index >= len(p.PyxisEventLogs) {
		err := fmt.Errorf("error. index %d out of range. Number of Pyxis event logs: %d", index, len(p.PyxisEventLogs))
		p.logger.LogError(fmt.Sprintf("cleanUpPyxisEventLog was called with an invalid index: %s", err.Error()))
		return err
	}

	//-- resort the events
	p.logger.LogInfo(fmt.Sprintf("sorting %s event log", p.PyxisEventLogs[index].PyxisName))
	sort.Slice(p.PyxisEventLogs[index].Log, func(i, j int) bool {
		return p.PyxisEventLogs[index].Log[i].TxDateTime.Before(p.PyxisEventLogs[index].Log[j].TxDateTime)
	})
	p.logger.LogInfo(fmt.Sprintf("%s sort complete", p.PyxisEventLogs[index].PyxisName))

	//-- check for duplicates
	p.logger.LogInfo(fmt.Sprintf("checking %s event log for duplicates", p.PyxisEventLogs[index].PyxisName))
	newLog := []PyxisEvent{}
	numDups := 0
	newLog = append(newLog, p.PyxisEventLogs[index].Log[0])
	for i := 1; i < len(p.PyxisEventLogs[index].Log); i++ {
		if p.PyxisEventLogs[index].Log[i] == p.PyxisEventLogs[index].Log[i-1] {
			numDups++
			continue
		} else {
			newLog = append(newLog, p.PyxisEventLogs[index].Log[i])
		}
	}
	p.PyxisEventLogs[index].Log = newLog
	switch numDups {
	case 0:
		p.logger.LogInfo("check complete. no duplicates found")

	case 1:
		p.logger.LogInfo("check complete. 1 duplicate removed")

	default:
		p.logger.LogInfo(fmt.Sprintf("check complete. %d duplicates removed", numDups))
	}

	//-- update date range
	oldDateTime := p.PyxisEventLogs[index].LastEventDateTime
	p.PyxisEventLogs[index].LastEventDateTime = p.PyxisEventLogs[index].Log[len(p.PyxisEventLogs[index].Log)-1].TxDateTime
	if p.PyxisEventLogs[index].LastEventDateTime.Compare(oldDateTime) != 0 {
		p.logger.LogInfo(fmt.Sprintf("%s last event updated from %s to %s",
			p.PyxisEventLogs[index].PyxisName,
			oldDateTime.Format("2006-01-02 1504"),
			p.PyxisEventLogs[index].LastEventDateTime.Format("2006-01-02 1504")))
	}

	return nil
}

func (p *ProcessState) addPyxisEvents(index int, events []PyxisEvent) error {
	if index >= len(p.PyxisEventLogs) {
		err := fmt.Errorf("error. index %d out of range. Number of Pyxis event logs: %d", index, len(p.PyxisEventLogs))
		p.logger.LogError(fmt.Sprintf("addPyxisEvents was called with an invalid index: %s", err.Error()))
		return err
	}
	p.logger.LogInfo(fmt.Sprintf("adding %d events to %s event log",
		len(events),
		p.PyxisEventLogs[index].PyxisName))

	p.PyxisEventLogs[index].Log = append(p.PyxisEventLogs[index].Log, events...)

	p.logger.LogInfo("events added")

	err := p.cleanUpPyxisEventLog(index)
	if err != nil {
		p.logger.LogError("Error calling cleanUpPyxisEventLog from addPyxisEvents. Pyxis event log may be out of order and/or contain duplicates")
		return err
	}

	return nil
}

func (p *PyxisEventLog) lastEventDateString() string {
	if p.LastEventDateTime.IsZero() {
		return ""
	}

	return p.LastEventDateTime.Format("2006-01-02 15:04")
}

func (p *ProcessState) parseEventsAndAdd(index int, events []database.PyxisEventResponse) {
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

	p.addPyxisEvents(index, parsedEvents)

}

func (p *ProcessState) createNewPyxisEventLog(pyxisName string, startDateTime time.Time) error {
	for _, pyxisLog := range p.PyxisEventLogs {
		if pyxisName == pyxisLog.PyxisName {
			err := fmt.Errorf("error. %s already exists", pyxisName)
			p.logger.LogError(fmt.Sprintf("Error. Failed to create new Pyxis event log: %s", err.Error()))
			return err
		}
	}

	p.PyxisEventLogs = append(p.PyxisEventLogs, PyxisEventLog{
		Log:           []PyxisEvent{},
		StartDateTime: startDateTime,
		PyxisName:     pyxisName,
	})
	p.logger.LogInfo(fmt.Sprintf("New Pyxis event log: %s added. Logging events starting on or after %s.",
		pyxisName,
		startDateTime.Format("2006-01-02 1504")))

	return nil
}

func (p *ProcessState) findMissingPyxisEvents() {
	for i := range p.PyxisEventLogs {
		startTime := time.Time{}
		if p.PyxisEventLogs[i].LastEventDateTime.IsZero() {
			startTime = p.PyxisEventLogs[i].StartDateTime
		} else {
			startTime = p.PyxisEventLogs[i].LastEventDateTime
		}

		endTime := timeToday() //--time today at midnight for cache if duplicate call

		if endTime.Sub(startTime) < minPyxisEventRecheckInterval {
			p.logger.LogInfo(fmt.Sprintf("Last Pyxis event for %s on %s, less than 24 hours ago. Finding missing Pyxis events skipped",
				p.PyxisEventLogs[i].PyxisName,
				startTime.Format("2006-01-02 1504")))
			continue
		}

		params := database.GetPyxisEventsForDeviceByDateRangeParams{
			Device: p.PyxisEventLogs[i].PyxisName,
			Start:  startTime,
			End:    endTime,
		}

		events, err := getPyxisEvents(p, params) //-- logging handled in function
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		p.parseEventsAndAdd(i, events)
	}
}

func (p *ProcessState) savePyxisEventLogs() error {
	for _, pyxisEventLog := range p.PyxisEventLogs {
		p.logger.LogInfo(fmt.Sprintf("Saving %s Pyxis event log", pyxisEventLog.PyxisName))
		logFile, err := os.OpenFile(filepath.Join(p.pathToData, pyxisEventLogsFolder, pyxisEventLog.PyxisName+".csv"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error opening %s Pyxis events: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}
		defer logFile.Close()

		err = gocsv.MarshalFile(pyxisEventLog.Log, logFile)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error saving %s Pyxis events: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}

		settings := struct {
			StartDateTime     time.Time
			LastEventDateTime time.Time
			PyxisName         string
		}{
			StartDateTime:     pyxisEventLog.StartDateTime,
			LastEventDateTime: pyxisEventLog.LastEventDateTime,
			PyxisName:         pyxisEventLog.PyxisName,
		}

		data, err := json.Marshal(&settings)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error marshalling log settings for %s Pyxis: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}

		saveFile, err := os.OpenFile(filepath.Join(p.pathToData, pyxisEventLogSettingsFolder, pyxisEventLog.PyxisName+".json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error opening %s Pyxis settings: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}
		defer saveFile.Close()

		_, err = saveFile.Write(data)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error saving %s Pyxis settings: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}
	}

	return nil
}

func (p *ProcessState) loadPyxisEventLogs() error {
	type unmatchedLog struct {
		Name string
		Logs []PyxisEvent
	}

	type logSettings struct {
		StartDateTime     time.Time
		LastEventDateTime time.Time
		PyxisName         string
	}

	p.logger.LogInfo("Loading Pyxis event logs")

	//-- Pull logs from csv files
	entries, err := os.ReadDir(filepath.Join(p.pathToData, pyxisEventLogsFolder))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error accessing Pyxis event save directory: %s", err.Error()))
		return err
	}

	unmatchedLogs := []unmatchedLog{}
	for _, entry := range entries {
		file, err := os.Open(filepath.Join(p.pathToData, pyxisEventLogsFolder, entry.Name()))
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error opening %s: %s", file.Name(), err.Error()))
			continue
		}
		defer file.Close()

		log := []PyxisEvent{}
		gocsv.UnmarshalFile(file, &log)
		unmatchedLogs = append(unmatchedLogs, unmatchedLog{
			Name: strings.Split(entry.Name(), ".")[0],
			Logs: log,
		})
	}

	//-- Pull settings from json files
	entries, err = os.ReadDir(filepath.Join(p.pathToData, pyxisEventLogSettingsFolder))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error accessing Pyxis event logs settings directory: %s", err.Error()))
		return err
	}

	unmatchedSettings := []logSettings{}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(p.pathToData, pyxisEventLogSettingsFolder, entry.Name()))
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error reading %s: %s", entry.Name(), err.Error()))
			continue
		}

		settings := logSettings{}
		err = json.Unmarshal(data, &settings)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error unmarshalling data from %s: %s", entry.Name(), err.Error()))
			continue
		}

		unmatchedSettings = append(unmatchedSettings, settings)
	}

	//-- Merge unmatchedLogs and unmatchedSettings
	pyxisEventLogs := []PyxisEventLog{}
	matchedLogs := []unmatchedLog{}
	matchedSettings := []logSettings{}

	for i := range unmatchedLogs {
		for j := range unmatchedSettings {
			if unmatchedLogs[i].Name == unmatchedSettings[j].PyxisName {
				pyxisEventLogs = append(pyxisEventLogs, PyxisEventLog{
					Log:               unmatchedLogs[i].Logs,
					StartDateTime:     unmatchedSettings[j].StartDateTime,
					LastEventDateTime: unmatchedSettings[j].LastEventDateTime,
					PyxisName:         unmatchedSettings[j].PyxisName,
				})

				matchedLogs = append(matchedLogs, unmatchedLogs[i])
				matchedSettings = append(matchedSettings, unmatchedSettings[j])

				break
			}
		}
	}

	//-- Check for unmatched logs and settings
	if len(matchedLogs) != len(unmatchedLogs) ||
		len(matchedSettings) != len(unmatchedSettings) {
		p.logger.LogError("Error matching Pyxis logs and settings")
	}

	p.PyxisEventLogs = pyxisEventLogs
	p.logger.LogInfo("Pyxis event logs loaded")
	return nil
}
