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
	ControlEventLog   *ControlEventLog
	StartDateTime     time.Time
	LastEventDateTime time.Time
	PyxisName         string
}

func (p *PyxisEventLog) CleanUp() *logResponder {
	logger := logResponder{}

	//-- resort the events
	logger.AddInfo(fmt.Sprintf("sorting %s event log", p.PyxisName))
	sort.Slice(p.Log, func(i, j int) bool {
		return p.Log[i].TxDateTime.Before(p.Log[j].TxDateTime)
	})
	logger.AddInfo(fmt.Sprintf("%s sort complete", p.PyxisName))

	//-- check for duplicates
	logger.AddInfo(fmt.Sprintf("checking %s event log for duplicates", p.PyxisName))
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
		logger.AddInfo("check complete. no duplicates found")

	case 1:
		logger.AddInfo("check complete. 1 duplicate removed")

	default:
		logger.AddInfo(fmt.Sprintf("check complete. %d duplicates removed", numDups))
	}

	//-- update date range
	oldDateTime := p.LastEventDateTime
	p.LastEventDateTime = p.Log[len(p.Log)-1].TxDateTime
	if p.LastEventDateTime.Compare(oldDateTime) != 0 {
		logger.AddInfo(fmt.Sprintf("%s last event updated from %s to %s",
			p.PyxisName,
			oldDateTime.Format("2006-01-02 1504"),
			p.LastEventDateTime.Format("2006-01-02 1504")))
	}

	return &logger
}

func (p *PyxisEventLog) AddPyxisEvents(events []PyxisEvent) *logResponder {
	logger := logResponder{}
	logger.AddInfo(fmt.Sprintf("adding %d events to %s event log",
		len(events),
		p.PyxisName))

	p.Log = append(p.Log, events...)
	logger.AddInfo("events added")

	logger.AddResponses(p.CleanUp())

	return &logger
}

func (p *PyxisEventLog) lastEventDateString() string {
	if p.LastEventDateTime.IsZero() {
		return ""
	}

	return p.LastEventDateTime.Format("2006-01-02 15:04")
}

func (p *PyxisEventLog) ParseEventsAndAdd(events []database.PyxisEventResponse) *logResponder {
	parsedEvents := []PyxisEvent{}

	for _, event := range events {
		pyxisEvent := PyxisEvent{}
		b, _ := event.ItemTransactionKey.MarshalBinary()
		pyxisEvent.ItemTransactionKey = uuid.UUID{
			b[3], b[2], b[1], b[0],
			b[5], b[4],
			b[7], b[6],
			b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15],
		}

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

	return p.AddPyxisEvents(parsedEvents)

}

func (p *PyxisEventLog) UnloadPyxisEvents() {
	p.Log = []PyxisEvent{}
}

func (p *PyxisEventLog) checkForNewControlEvents() *logResponder {
	logger := logResponder{}
	logger.AddInfo(fmt.Sprintf("Checking for new control events for %s Pyxis", p.PyxisName))

	//-- Define med class codes that are controlled in map for checking
	controlClassCodes := map[string]struct{}{
		"2": struct{}{},
		"3": struct{}{},
		"4": struct{}{},
		"5": struct{}{},
	}

	//-- Create slice of control pyxis events
	controlEvents := []PyxisEvent{}
	for _, event := range p.Log {
		switch event.TransactionType {
		case "Remove":
			fallthrough
		case "Waste":
			fallthrough
		case "IntWaste":
			fallthrough
		case "Return to bin":
			if _, okay := controlClassCodes[event.MedClassCode]; okay {
				controlEvents = append(controlEvents, event)
			}
		}
	}

	unmatchedEvents := []PyxisEvent{}
	loggedControlEvents := p.ControlEventLog.GetLoggedControlEventKeys()

	for _, controlEvent := range controlEvents {
		if _, okay := loggedControlEvents[controlEvent.ItemTransactionKey]; !okay {
			unmatchedEvents = append(unmatchedEvents, controlEvent)
		}
	}

	if len(unmatchedEvents) == 0 {
		logger.AddInfo("No new control events found")
	} else {
		logger.AddInfo(fmt.Sprintf("%d new control events found. Adding to unmatched control log", len(unmatchedEvents)))
	}

	p.ControlEventLog.UnmatchedEvents = unmatchedEvents
	return &logger

}

func (p *Process) saveAndUnloadPyxisEventLogs() error {
	p.logger.LogInfo("Saving pyxis event logs")
	for i, pyxisEventLog := range p.PyxisEventLogs {
		//-- Save control event log as these stay loaded
		p.logger.LogInfo(fmt.Sprintf("Saving control event log for %s", pyxisEventLog.PyxisName))
		if err := p.PyxisEventLogs[i].ControlEventLog.Save(p); err != nil {
			p.logger.LogError("Error. Save failed")
			return err
		}
		//-- Skip to next pyxis if current is not loaded
		if !p.state.IsLoaded(pyxisEventLog.PyxisName) {
			continue
		}

		//-- Marshal and write pyxis event log data to csv file
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

		//-- Marshall and write pyxis event log settings data
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

		//-- Marshall and write control event log data
		data, err = json.Marshal(&pyxisEventLog.ControlEventLog)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error marshalling control event log for %s Pyxis: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}

		controlFile, err := os.OpenFile(filepath.Join(p.pathToData, controlEventLogsFolder, pyxisEventLog.PyxisName+".json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error opening %s control event log: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}
		defer controlFile.Close()

		_, err = controlFile.Write(data)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error saving %s control event log: %s", pyxisEventLog.PyxisName, err.Error()))
			return err
		}

		//-- Remove pyxis from list of loaded event logs
		p.logger.LogInfo(fmt.Sprintf("%s pyxis event log saved", pyxisEventLog.PyxisName))
		pyxisEventLog.UnloadPyxisEvents()
		p.state.PyxisEventLogUnloaded(pyxisEventLog.PyxisName)
		p.logger.LogInfo(fmt.Sprintf("%s pyxis event log unloaded", pyxisEventLog.PyxisName))
	}

	return nil
}

func (p *Process) loadPyxisEventlog(pyxis string) error {
	type logSettings struct {
		StartDateTime     time.Time
		LastEventDateTime time.Time
		PyxisName         string
	}

	file, err := os.Open(filepath.Join(p.pathToData, pyxisEventLogsFolder, pyxis+".csv"))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to load %s Pyxis event log: %s", pyxis, err.Error()))
		return err
	}
	defer file.Close()

	log := []PyxisEvent{}
	gocsv.UnmarshalFile(file, &log)

	data, err := os.ReadFile(filepath.Join(p.pathToData, pyxisEventLogSettingsFolder, pyxis+".json"))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to read %s Pyxis event log settings: %s", pyxis, err.Error()))
		return err
	}

	settings := logSettings{}
	err = json.Unmarshal(data, &settings)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error unmarshalling settings data for %s: %s", pyxis, err.Error()))
		return err
	}

	data, err = os.ReadFile(filepath.Join(p.pathToData, controlEventLogsFolder, pyxis+".json"))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to read %s control event log file: %s", pyxis, err.Error()))
		return err
	}

	controlLog := ControlEventLog{}
	err = json.Unmarshal(data, &controlLog)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error unmarshalling control event log data for %s: %s", pyxis, err.Error()))
		return err
	}

	pyxisEventLog := PyxisEventLog{
		Log:               log,
		ControlEventLog:   &controlLog,
		StartDateTime:     settings.StartDateTime,
		LastEventDateTime: settings.LastEventDateTime,
		PyxisName:         pyxis,
	}

	pyxisEventLog.ControlEventLog.pyxisEventLog = &pyxisEventLog

	p.PyxisEventLogs = append(p.PyxisEventLogs, &pyxisEventLog)
	p.state.PyxisEventLogLoaded(pyxis)
	return nil

}

func (p *Process) loadPyxisEventLogs() error {
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
	pyxisEventLogs := []*PyxisEventLog{}
	matchedLogs := []unmatchedLog{}
	matchedSettings := []logSettings{}

	for i := range unmatchedLogs {
		for j := range unmatchedSettings {
			if unmatchedLogs[i].Name == unmatchedSettings[j].PyxisName {
				pyxisEventLogs = append(pyxisEventLogs, &PyxisEventLog{
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
		return fmt.Errorf("error. unable to match pyxis logs and settings on load")
	}

	p.PyxisEventLogs = pyxisEventLogs

	//-- Pull control event logs
	for i, pyxisEventLog := range p.PyxisEventLogs {
		//-- Check if control event log file exists. Add if doesn't exist
		if _, err := os.Stat(filepath.Join(p.pathToData, controlEventLogsFolder, pyxisEventLog.PyxisName+".json")); err != nil {
			if os.IsNotExist(err) {
				p.PyxisEventLogs[i].ControlEventLog = &ControlEventLog{
					Log:             []ControlEventTrail{},
					UnmatchedEvents: []PyxisEvent{},
					pyxisEventLog:   p.PyxisEventLogs[i],
				}
				pyxisEventLogs = append(pyxisEventLogs, pyxisEventLog)
				continue
			} else {
				p.logger.LogError(fmt.Sprintf("Error. Unable to read control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName))
				continue
			}
		}

		//-- Read data from control event log file
		data, err := os.ReadFile(filepath.Join(p.pathToData, controlEventLogsFolder, pyxisEventLog.PyxisName+".json"))
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error. Unable to read control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName))
			continue
		}

		err = json.Unmarshal(data, &p.PyxisEventLogs[i].ControlEventLog)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Error. Unable to unmarshall control event log for %s. Pyxis event log not loaded.", pyxisEventLog.PyxisName))
			continue
		}

		//-- Link PyxisEventLog pointer in ControlEventLog
		p.PyxisEventLogs[i].ControlEventLog.pyxisEventLog = p.PyxisEventLogs[i]

		//-- Add pyxis name to list of loaded pyxis event logs in process state
		logErr := p.state.PyxisEventLogLoaded(pyxisEventLog.PyxisName)
		if logErr != nil {
			p.logger.LogError(logErr.logMessage)
			return logErr
		}
	}

	p.state.PyxisEventLogsLoadSuccessful()
	p.logger.LogInfo("Pyxis event logs loaded")
	return nil
}
