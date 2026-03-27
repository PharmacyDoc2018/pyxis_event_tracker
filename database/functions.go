package database

import (
	"fmt"
	"time"
)

func parseDate(dateString string) (time.Time, error) {
	dateFormats := []string{
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
