package main

import (
	"fmt"
	"strconv"
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

func timeToday() time.Time {
	now := time.Now()
	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		now.Location(),
	)

	return today
}

func timeStartDay(t time.Time) time.Time {
	startDay := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		t.Location(),
	)

	return startDay
}

func timeEndDay(t time.Time) time.Time {
	startDay := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		23,
		59,
		59,
		0,
		t.Location(),
	)

	return startDay
}

func isSameDay(timeOne, timeTwo time.Time) bool {
	return timeOne.Year() == timeTwo.Year() &&
		timeOne.Month() == timeTwo.Month() &&
		timeOne.Day() == timeTwo.Day()
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func addFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt + ymInt
	return float64(zmInt) / m

}

func subtractFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt - ymInt
	return float64(zmInt) / m

}

func quickDisplayName(f func(string) (string, *logError), s string) string {
	r, e := f(s)
	if e != nil {
		return "unknown"
	}
	return r
}

func printfln(format string, a ...any) (n int, err error) {
	n, err = fmt.Println(fmt.Sprintf(format, a))
	return n, err
}
