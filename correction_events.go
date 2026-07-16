package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

type CorrectionEventLink struct {
	Id          string
	WriteUpFile string
}

type CorrectionEvent struct {
	Id             string
	EventDate      time.Time
	CorrectionDate time.Time
	UserID         string
	UserName       string
	ItemId         string
	DisplayName    string
	Amount         float64
	Units          string
	MRN            string
	PtName         string
	BeSafe         string
}

const CorrectionEventsDirName = "correction_events"
const CorrectionWriteUpsDirName = "write_ups"
const CorrectionEventLinksFileName = "correction_event_links.json"

type CorrectionEventLinks struct {
	//-- [dateID][index]CorrectionEvent
	Map map[string]map[int]CorrectionEventLink
}

func (c *CorrectionEventLinks) AddAndLink(p *Process, pyxisName string, mrn, itemID string, date time.Time, newCorrectionEvent CorrectionEvent) *logError {
	dateID := timeStartDay(date).Format("20060102")
	if _, okay := c.Map[dateID]; !okay {
		c.Map[dateID] = map[int]CorrectionEventLink{}
	}

	index := 0
	_, okay := c.Map[dateID][index]
	for okay {
		index++
		_, okay = c.Map[dateID][index]
	}

	indexName := ""
	if index <= 9 {
		indexName = "0" + strconv.Itoa(index)
	} else {
		indexName = strconv.Itoa(index)
	}

	id := dateID + indexName
	fileName := fmt.Sprintf("%s.txt", id)

	newCorrectionEvent.Id = id

	link := CorrectionEventLink{
		Id:          id,
		WriteUpFile: filepath.Join(p.pathToData, CorrectionEventsDirName, CorrectionWriteUpsDirName, fileName),
	}

	logIndex := 0
	found := false
	for i := range p.PyxisEventLogs {
		if p.PyxisEventLogs[i].PyxisName == pyxisName {
			found = true
			logIndex = i
		}
	}
	if !found {
		return &logError{
			logMessage: fmt.Sprintf("Error. %s Pyxis not found", pyxisName),
			errMessage: fmt.Sprintf("error. %s pyxis not found", pyxisName),
		}
	}

	items := []EventTrailItem{}
	for item := range p.selectedEventActions.Map {
		items = append(items, item)
	}

	items = append(items, EventTrailItem{
		Type:            correctionEvent,
		CorrectionEvent: newCorrectionEvent,
	})

	sort.Slice(items, func(i, j int) bool {
		dateTimeOne := time.Time{}
		switch items[i].Type {
		case pyxisEvent:
			dateTimeOne = items[i].PyxisEvent.TxDateTime

		case marAction:
			dateTimeOne = items[i].MarAction.SavedTime

		case correctionEvent:
			dateTimeOne = items[i].CorrectionEvent.CorrectionDate
		}

		dateTimeTwo := time.Time{}
		switch items[j].Type {
		case pyxisEvent:
			dateTimeTwo = items[j].PyxisEvent.TxDateTime

		case marAction:
			dateTimeTwo = items[j].MarAction.SavedTime

		case correctionEvent:
			dateTimeTwo = items[j].CorrectionEvent.CorrectionDate
		}

		return dateTimeOne.Before(dateTimeTwo)
	})

	logErr := p.PyxisEventLogs[logIndex].ControlEventLog.LinkEventActions(mrn, itemID, date, items...)
	if logErr != nil {
		//-- Remove new correction action from selected events
		for item := range p.selectedEventActions.Map {
			if item.Type == correctionEvent {
				delete(p.selectedEventActions.Map, item)
			}
		}
		return logErr
	}

	c.Map[dateID][index] = link
	return nil

}

func (c *CorrectionEventLinks) Load(dataPath string) (error, *logResponder) {
	logger := logResponder{}
	logger.AddInfo("Loading correction events")

	_, err := os.Stat(filepath.Join(dataPath, CorrectionEventsDirName))
	if err != nil {
		if os.IsNotExist(err) {
			logger.AddInfo("correction_event directory not found. Attempting to create directory")
			err = os.Mkdir(filepath.Join(dataPath, CorrectionEventsDirName), 0755)
			if err != nil {
				logger.AddError(fmt.Sprintf("Load failed. %s", err.Error()))
				return err, &logger
			}

			logger.AddInfo("correction_event directory successfully created. Attempting to create Write_ups directory")
			err = os.Mkdir(filepath.Join(dataPath, CorrectionEventsDirName, CorrectionWriteUpsDirName), 0755)
			if err != nil {
				logger.AddError(fmt.Sprintf("Load failed. %s", err.Error()))
				return err, &logger
			}

			logger.AddInfo("write_ups directory successfully created. Attempting to create correction_event_links file")
			file, err := os.OpenFile(filepath.Join(dataPath, CorrectionEventsDirName, CorrectionEventLinksFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				logger.AddError(fmt.Sprintf("Load failed. %s", err.Error()))
				return err, &logger
			}
			_, err = file.WriteString("{}")
			if err != nil {
				logger.AddError(fmt.Sprintf("Load failed. %s", err.Error()))
				return err, &logger
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(dataPath, CorrectionEventsDirName, CorrectionEventLinksFileName))
	if err != nil {
		logger.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &logger
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, &c)
	if err != nil {
		logger.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &logger

	}

	logger.AddInfo("Correction events loaded successfully")
	return nil, &logger

}

func (c *CorrectionEventLinks) Save(dataPath string) (error, *logResponder) {
	logger := logResponder{}
	logger.AddInfo("Saving correction event links")

	file, err := os.OpenFile(filepath.Join(dataPath, CorrectionEventsDirName, CorrectionEventLinksFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error. Unable to open %s: %s", CorrectionEventLinksFileName, err.Error()))
		return err, &logger
	}
	defer file.Close()

	data, err := json.Marshal(&c)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error marshalling correction event links: %s", err.Error()))
		return err, &logger
	}

	_, err = file.Write(data)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error writting marshalled correction event links to file %s: %s", ERxFileName, err.Error()))
		return err, &logger
	}

	return nil, &logger

}

func initCorrectionEvents() *CorrectionEventLinks {
	linkMap := map[string]map[int]CorrectionEventLink{}
	return &CorrectionEventLinks{
		Map: linkMap,
	}
}
