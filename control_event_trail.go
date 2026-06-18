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

func (c *ControlEventLog) SortUnmatchedEvents() {
	sort.Slice(c.UnmatchedEvents, func(i, j int) bool {
		return c.UnmatchedEvents[i].TxDateTime.Before(c.UnmatchedEvents[j].TxDateTime)
	})
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

func (c *ControlEventLog) MatchEvents(pyxisEvents []PyxisEvent, marActions []MarAction, date time.Time, mrn, itemID string) []PyxisEvent {
	allEvents := []EventTrailItem{}

	for _, event := range pyxisEvents {
		allEvents = append(allEvents, EventTrailItem{
			Type:       pyxisEvent,
			PyxisEvent: event,
		})
	}

	for _, action := range marActions {
		allEvents = append(allEvents, EventTrailItem{
			Type:      marAction,
			MarAction: action,
		})
	}

	sort.Slice(allEvents, func(i, j int) bool {
		timeOne := time.Time{}
		timeTwo := time.Time{}

		if allEvents[i].Type == pyxisEvent {
			timeOne = allEvents[i].PyxisEvent.TxDateTime
		} else {
			timeOne = allEvents[i].MarAction.SavedTime
		}

		if allEvents[j].Type == pyxisEvent {
			timeTwo = allEvents[j].PyxisEvent.TxDateTime
		} else {
			timeTwo = allEvents[j].MarAction.SavedTime
		}

		return timeOne.Before(timeTwo)
	})

	controlEventTrail := ControlEventTrail{
		MRN:         mrn,
		ItemID:      itemID,
		Date:        date,
		EventTrails: []EventTrail{},
		Vaild:       false,
	}

	trackedTransactions := map[string]struct{}{
		"IntWaste":         struct{}{},
		"Remove":           struct{}{},
		"Remove Cancelled": struct{}{},
		"Return to bin":    struct{}{},
		"Waste":            struct{}{},
	}

	matchStatus := struct {
		InitialRemoveIndex     int
		InitialRemoveAmount    float64
		SubsequentRemoveIndexs []int
		CurrentAmount          float64
		Done                   bool
	}{
		InitialRemoveIndex:     0,
		InitialRemoveAmount:    0.0,
		SubsequentRemoveIndexs: []int{},
		CurrentAmount:          0.0,
		Done:                   false,
	}

	for !matchStatus.Done {
		for i, event := range allEvents {
			switch event.Type {
			case pyxisEvent:
				switch event.PyxisEvent.TransactionType {
				case "Remove":
					if matchStatus.InitialRemoveAmount == 0.0 {
						matchStatus.InitialRemoveAmount = event.PyxisEvent.AmountReferenced
						matchStatus.InitialRemoveIndex = i
						matchStatus.CurrentAmount = event.PyxisEvent.AmountReferenced
					} else {
						matchStatus.CurrentAmount = addFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)
						matchStatus.SubsequentRemoveIndexs = append(matchStatus.SubsequentRemoveIndexs, i)
					}
				case "Waste":
					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)

				case "IntWaste":
					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)

				case "Return to bin":
					matchStatus.CurrentAmount = subtractFloat(matchStatus.CurrentAmount, event.PyxisEvent.AmountReferenced)
				}

			case marAction:
				matchStatus.CurrentAmount = addFloat(matchStatus.CurrentAmount, event.MarAction.CalcMinDose)
			}

			//-- Check for trail end:
			//-- If current amount = 0 and there is at least one remove event -> end trail
			if matchStatus.CurrentAmount == 0.0 && matchStatus.InitialRemoveAmount > 0.0 {
				newTrail := allEvents[0 : i+1]
				if len(newTrail) == len(allEvents) {
					allEvents = []EventTrailItem{}
					matchStatus.Done = true
				} else {
					allEvents = allEvents[i+2:]
				}

				tempTrail := newTrail
				newTrail = []EventTrailItem{}
				for _, item := range tempTrail {
					switch item.Type {
					case pyxisEvent:
						if _, okay := trackedTransactions[item.PyxisEvent.TransactionType]; okay {
							newTrail = append(newTrail, item)
						}

					case marAction:
						newTrail = append(newTrail, item)
					}
				}

				controlEventTrail.EventTrails = append(controlEventTrail.EventTrails, EventTrail{
					Trail: newTrail,
					Vaild: true,
				})
				break
			}

			//-- if current amount == initial remove amount AND there are more than one remove events -> piece trail together
			if matchStatus.CurrentAmount == matchStatus.InitialRemoveAmount && len(matchStatus.SubsequentRemoveIndexs) > 0 {
				removeIndexsToRemove := map[int]struct{}{}
				amountSoFar := 0.0
				for j, item := range allEvents[0 : i+1] {
					if item.PyxisEvent.TransactionType == "Remove" &&
						addFloat(item.PyxisEvent.AmountReferenced, amountSoFar) <= matchStatus.CurrentAmount {
						amountSoFar = addFloat(amountSoFar, item.PyxisEvent.AmountReferenced)
						removeIndexsToRemove[j] = struct{}{}
					}
				}

				tempTrail := allEvents[0 : i+1]
				newTrail := []EventTrailItem{}
				addBackTrail := []EventTrailItem{}

				for j, item := range tempTrail {
					switch item.Type {
					case pyxisEvent:
						if _, okay := removeIndexsToRemove[j]; okay {
							addBackTrail = append(addBackTrail, item)
							continue
						}
						if _, okay := trackedTransactions[item.PyxisEvent.TransactionType]; okay {
							newTrail = append(newTrail, item)
						}

					case marAction:
						newTrail = append(newTrail, item)
					}
				}

				controlEventTrail.EventTrails = append(controlEventTrail.EventTrails, EventTrail{
					Trail: newTrail,
					Vaild: true,
				})

				if len(tempTrail) == len(allEvents) && len(addBackTrail) == 0 {
					allEvents = []EventTrailItem{}
					matchStatus.Done = true
				} else {
					allEvents = allEvents[i+2:]
					allEvents = append(addBackTrail, allEvents...)
				}
				break
			}

			if i+1 >= len(allEvents) {
				matchStatus.Done = true
			}
		}

	}
	if len(controlEventTrail.EventTrails) > 0 {
		c.Log = append(c.Log, controlEventTrail)
	}
	unmatchedEvents := []PyxisEvent{}
	for _, event := range allEvents {
		if event.Type == pyxisEvent {
			unmatchedEvents = append(unmatchedEvents, event.PyxisEvent)
		}
	}
	c.UnmatchedEvents = append(c.UnmatchedEvents, unmatchedEvents...)
	return unmatchedEvents

}
