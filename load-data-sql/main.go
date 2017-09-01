package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	marsimages "github.com/danisla/go-marsimages"
	_ "github.com/go-sql-driver/mysql"
)

func importImages(solStart, solEnd int, db *sql.DB) (int, error) {

	const manifestURL = "https://mars.jpl.nasa.gov/msl-raw-images/image/image_manifest.json"

	manifest, err := marsimages.FetchManifest(manifestURL)
	if err != nil {
		return 0, err
	}

	dbInsert, err := db.Prepare("INSERT IGNORE INTO images VALUES( ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer dbInsert.Close()

	var wg sync.WaitGroup

	total := len(manifest.Sols)
	var count uint64

	rangeStart := solStart
	if solStart < 0 {
		rangeStart = total + solStart
	}

	rangeEnd := solEnd
	if solEnd < 0 {
		rangeEnd = total + solEnd
	}

	for i := rangeStart; i <= rangeEnd; i++ {
		wg.Add(1)
		solImages := 0
		url := manifest.Sols[i].CatalogURL
		go func(url string) {
			defer wg.Done()
			catalog, err := marsimages.FetchCatalog(url)
			if err != nil {
				log.Printf("Error fetching url: %s: %s\n", url, err)
			} else {
				for j := 0; j < len(catalog.Images); j++ {
					img := catalog.Images[j]
					if img.SampleType != "thumbnail" {
						solImages++
						sol, err := strconv.ParseInt(img.Sol, 10, 64)
						if err != nil {
							sol = -1
						}

						// Convert http urls to https, avoids redirect when loading.
						httpsURL := strings.Replace(img.URL, "http://", "https://", -1)

						marsImage := marsimages.MarsImage{ItemName: img.ItemName, URL: httpsURL, Instrument: img.Instrument, LMST: img.LMST, Sol: sol, UTC: img.UTC}

						wg.Add(1)

						go func(marsImage marsimages.MarsImage) {
							defer wg.Done()
							_, err := dbInsert.Exec(marsImage.Instrument, marsImage.ItemName, marsImage.LMST, marsImage.Sol, marsImage.URL, marsImage.UTC)
							if err != nil {
								log.Printf("Error inserting image: %s. %v", marsImage.URL, err)
							} else {
								atomic.AddUint64(&count, 1)
							}
						}(marsImage)

					}
				}
				log.Printf("Found %d/%d full scale images for sol %d", solImages, len(catalog.Images), catalog.Sol)
			}
		}(url)
	}
	wg.Wait()

	return int(atomic.LoadUint64(&count)), nil
}

func main() {
	var connectionName string
	var user string
	var password string
	var database string
	var solStart int
	var solEnd int
	var dropDB bool

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

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, connectionName, database))
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
