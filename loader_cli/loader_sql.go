package main

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

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
