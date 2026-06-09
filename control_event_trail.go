package main

import "time"

type EventTrail struct {
	Trail []int
	Vaild bool
}

type ControlEventTrail struct {
	MRN         string
	ItemID      string
	Date        time.Time
	PyxisEvents []PyxisEvent
	AdminEvents []MarEvent
	EventTrails []EventTrail
	Vaild       bool
}
