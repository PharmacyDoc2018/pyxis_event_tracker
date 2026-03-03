package main

import (
	"context"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	c := initConfig()
	defer c.db.Close()

	cTwoMedList, err := c.dbq.ListControlTwoMedsByDevice(context.Background(), "AUGUSTA2")
	if err != nil {
		fmt.Printf("Failed to retrieve C2 list from db: %s", err.Error())
	} else {
		fmt.Println("Medications found! Printing list:")
	}

	for _, med := range cTwoMedList {
		fmt.Println(med)
	}
}
