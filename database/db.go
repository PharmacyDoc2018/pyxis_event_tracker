package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func ParseDose(dispQty float64, unitName, strength string) (float64, error) {
	doseRunes := map[rune]struct{}{
		'0': struct{}{},
		'1': struct{}{},
		'2': struct{}{},
		'3': struct{}{},
		'4': struct{}{},
		'5': struct{}{},
		'6': struct{}{},
		'7': struct{}{},
		'8': struct{}{},
		'9': struct{}{},
		'.': struct{}{},
		',': struct{}{},
	}

	strFloat := 0.0
	var err error

	splitStrength := strings.Split(strength, "/")
	switch len(splitStrength) {
	case 0:
		return 0.0, fmt.Errorf("error. invalid strength format")

	//-- Not a concentration
	case 1:
		//-- Get the number from the start of the strength string
		strRunes := []rune(strength)
		i := 0
		_, okay := doseRunes[strRunes[i]]
		for okay {
			i++
			_, okay = doseRunes[strRunes[i]]
		}
		strRunes = strRunes[:i]

		//-- Filter out commas
		tempRunes := strRunes
		strRunes = []rune{}
		for i := range tempRunes {
			if tempRunes[i] != ',' {
				strRunes = append(strRunes, tempRunes[i])
			}
		}

		strFloat, err = strconv.ParseFloat(string(strRunes), 64)
		if err != nil {
			return 0.0, err
		}

	case 2:
		//
		strRunes := []rune(splitStrength[0])
		denominator := strings.TrimSpace(strings.ToLower(splitStrength[1]))
		if denominator != strings.TrimSpace(strings.ToLower(unitName)) {
			return 0.0, fmt.Errorf("error. %s in concentration %s does not match disp_qty unit %s", denominator, strength, unitName)
		}

		i := 0
		_, okay := doseRunes[strRunes[i]]
		for okay {
			i++
			_, okay = doseRunes[strRunes[i]]
		}
		strRunes = strRunes[:i]

		//-- Filter out commas
		tempRunes := strRunes
		strRunes = []rune{}
		for i := range tempRunes {
			if tempRunes[i] != ',' {
				strRunes = append(strRunes, tempRunes[i])
			}
		}

		strFloat, err = strconv.ParseFloat(string(strRunes), 64)
		if err != nil {
			return 0.0, err
		}

	default:
		return 0.0, fmt.Errorf("error. invalid strength format")
	}

	return multiplyFloat(dispQty, strFloat), nil
}

func multiplyFloat(x, y float64) float64 {
	m := 10000.0

	xmInt := int(x * m)
	ymInt := int(y * m)

	zmInt := xmInt * ymInt
	return float64(zmInt) / (m * m)

}
