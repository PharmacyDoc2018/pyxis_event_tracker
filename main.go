package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/microsoft/go-mssqldb"
)

func main() {
	godotenv.Load(".env")
	connString := os.Getenv("CONNSTRING")

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		fmt.Printf("Error creating connection pool: %s\n ", err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Printf("Error connecting to database:%s\n ", err.Error())
	}

	fmt.Println("Connected!")

	fmt.Println("Attempting to retrieve list of C2 medications from AUGUSTA2...")
	q := New(db)
	cTwoMedList, err := q.ListControlTwoMedsByDevice(context.Background(), "AUGUSTA2")
	if err != nil {
		fmt.Printf("Failed to retrieve C2 list from db: %s", err.Error())
	} else {
		fmt.Println("Medications found! Printing list:")
	}

	for _, med := range cTwoMedList {
		fmt.Println(med)
	}
}
