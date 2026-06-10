package main

import (
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
	Log []ControlEventTrail
}

func (c *ControlEventLog) Sort() {
	sort.Slice(c.Log, func(i, j int) bool {
		return c.Log[i].Date.Before(c.Log[j].Date)
	})
}

func (c *ControlEventLog) GetLoggedPyxisEventKeys() map[uuid.UUID]struct{} {
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
