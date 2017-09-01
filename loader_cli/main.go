package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var connectionProto string
	var connectionName string
	var user string
	var password string
	var database string
	var solStart int
	var solEnd int
	var dropDB bool

	flag.StringVar(&connectionProto, "proto", "tcp", "SQL connection protocol")
	flag.StringVar(&connectionName, "connection", "127.0.0.1:3306", "SQL host:port")
	flag.StringVar(&user, "user", "", "CloudSQL user name")
	flag.StringVar(&password, "password", "", "CloudSQL password, default is ''")
	flag.StringVar(&database, "database", "mars-images", "CloudSQL database name")
	flag.IntVar(&solStart, "start", -10, "Sol number to start scrape at, negative values index backwards from the latest sol")
	flag.IntVar(&solEnd, "end", -1, "Sol number to end scrape at, negative values index backwards from the latest sol and -1 indicates the latest sol.")
	flag.BoolVar(&dropDB, "drop", false, "Drop the table before upserting values")
	flag.Parse()

	if connectionName == "" {
		log.Fatal("Missing flag: -connection")
	}

	if user == "" {
		log.Fatal("Missing flag: -user")
	}

	if database == "" {
		log.Fatal("Missing flag: -database")
	}

	var err error
	var rows *sql.Rows
	var count int

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s)/%s", user, password, connectionProto, connectionName, database))
	if err != nil {
		log.Fatalf("Could not open db: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(runtime.NumCPU() * 2)

	if dropDB {
		log.Println("Droping existing table")

		// Clear the existing table
		rows, err = db.Query("DROP TABLE IF EXISTS images")
		if err != nil {
			log.Fatalf("Could not query db: %v", err)
			return
		}
		rows.Close()

		// Create the table
		log.Println("Creating table")

		rows, err = db.Query("CREATE TABLE images (instrument VARCHAR(32), itemName VARCHAR(100) PRIMARY KEY, lmst VARCHAR(32), sol int, url VARCHAR(255), utc DATETIME, INDEX items USING BTREE (itemName), INDEX utc USING BTREE (utc ASC), INDEX sol USING BTREE (sol ASC))")
		if err != nil {
			log.Fatalf("Could not query db: %v", err)
			return
		}
		rows.Close()
	}

	startTime := time.Now()

	count, err = importImages(solStart, solEnd, db)
	if err != nil {
		log.Fatalf("Error loading images: %v", err)
	}

	duration := time.Since(startTime)

	log.Printf("Imported %d images in: %s", count, duration)
}
