package main

import (
	"fmt"
	"time"
)

func parseDate(dateString string) (time.Time, error) {
	dateFormats := []string{
		"2006-01-02 15:04",
		"01/02/2006 15:04",
		"1/2/2006 15:04",
		"01/02/06 15:04",
		"1/2/06 15:04",
		"01-02-2006 15:04",
		"1-2-2006 15:04",
		"01-02-06 15:04",
		"1-2-06 15:04",
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"01/02/06",
		"1/2/06",
		"01-02-2006",
		"1-2-2006",
		"01-02-06",
		"1-2-06",
	}

	for _, format := range dateFormats {
		date, err := time.Parse(format, dateString)
		if err == nil {
			return date, nil
		}
	}

	return time.Time{}, fmt.Errorf("error. %s not a valid date format", dateString)
}
