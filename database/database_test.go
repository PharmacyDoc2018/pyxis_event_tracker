package database

import (
	"testing"
)

func TestParseDate(t *testing.T) {
	dateString := "01/01/2026"

	date, err := parseDate(dateString)
	if err != nil {
		t.Errorf("error. couldn't parse date: %s", err.Error())
	}

	newDateString := date.Format("2006-01-02")
	if newDateString != "2026-01-01" {
		t.Errorf("error. date not parsed correctly. expected 2026-01-01. actual %s", newDateString)
	}
}
