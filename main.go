package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
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
}
