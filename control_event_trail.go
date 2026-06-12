package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
)

type EventType int

const (
	pyxisEvent = iota
	marAction
)

type EventTrailItem struct {
	Type       EventType
	PyxisEvent PyxisEvent
	MarAction  MarAction
}

type EventTrail struct {
	Trail []EventTrailItem
	Vaild bool
}

type ControlEventTrail struct {
	MRN         string
	ItemID      string
	Date        time.Time
	EventTrails []EventTrail
	Vaild       bool
}

type ControlEventLog struct {
	Log             []ControlEventTrail
	UnmatchedEvents []PyxisEvent
	pyxisEventLog   *PyxisEventLog
}

func (c *ControlEventLog) Sort() {
	sort.Slice(c.Log, func(i, j int) bool {
		return c.Log[i].Date.Before(c.Log[j].Date)
	})
}

func (c *ControlEventLog) GetLoggedControlEventKeys() map[uuid.UUID]struct{} {
	eventKeys := map[uuid.UUID]struct{}{}

	for _, controlEventTrail := range c.Log {
		for _, eventTrail := range controlEventTrail.EventTrails {
			for _, item := range eventTrail.Trail {
				if item.Type == pyxisEvent {
					eventKeys[item.PyxisEvent.ItemTransactionKey] = struct{}{}
				}
			}
		}
	}

	return eventKeys
}

func (c *ControlEventLog) AddUnmatchedEvents(events []PyxisEvent) {
	c.UnmatchedEvents = append(c.UnmatchedEvents, events...)
}

func (c *ControlEventLog) Save(p *Process) error {
	//-- Marshall and write control event log data
	data, err := json.Marshal(&c)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling control event log for %s Pyxis: %s", c.pyxisEventLog.PyxisName, err.Error()))
		return err
	}

	controlFile, err := os.OpenFile(filepath.Join(p.pathToData, controlEventLogsFolder, c.pyxisEventLog.PyxisName+".json"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error opening %s control event log: %s", c.pyxisEventLog.PyxisName, err.Error()))
		return err
	}
	defer controlFile.Close()

	_, err = controlFile.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error saving %s control event log: %s", c.pyxisEventLog.PyxisName, err.Error()))
		return err
	}

	return nil
}
