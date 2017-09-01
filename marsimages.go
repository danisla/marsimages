package marsimages

import (
	"database/sql"
	"fmt"
	"html/template"
	golog "log"
	"math"
	"net/http"
	"os"
	"time"

	marsimages "github.com/danisla/go-marsimages"
	_ "github.com/go-sql-driver/mysql"
)

type pageData struct {
	Latest  marsimages.ListOfImages
	Mahli   marsimages.ListOfImages
	Mastcam marsimages.ListOfImages
	Navcam  marsimages.ListOfImages
	Hazcam  marsimages.ListOfImages
	Chemcam marsimages.ListOfImages
}

func init() {
	http.HandleFunc("/", imagesHandler)
}

func imagesHandler(w http.ResponseWriter, r *http.Request) {
	connectionProto := mustGetenv("SQL_CONNECTION_PROTO")
	connectionName := mustGetenv("CLOUDSQL_CONNECTION_NAME")
	user := mustGetenv("CLOUDSQL_USER")
	password := os.Getenv("CLOUDSQL_PASSWORD") // NOTE: password may be empty
	imageDb := mustGetenv("IMAGE_DATABASE")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s)/%s", user, password, connectionProto, connectionName, imageDb))
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not open db: %v", err), 500)
		return
	}
	defer db.Close()

	const limit int64 = 64

	var data pageData

	latest, err := queryImages(fmt.Sprintf("SELECT * FROM images ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	mahli, err := queryImages(fmt.Sprintf("SELECT * FROM images WHERE instrument IN ('MAHLI') ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	mastcam, err := queryImages(fmt.Sprintf("SELECT * FROM images WHERE instrument IN ('MAST_LEFT','MAST_RIGHT') ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	navcam, err := queryImages(fmt.Sprintf("SELECT * FROM images WHERE instrument IN ('NAV_LEFT_A','NAV_LEFT_B','NAV_RIGHT_A','NAV_RIGHT_B') ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	hazcam, err := queryImages(fmt.Sprintf("SELECT * FROM images WHERE instrument IN ('RHAZ_LEFT_A','RHAZ_LEFT_B','RHAZ_RIGHT_A','RHAZ_RIGHT_B') ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	chemcam, err := queryImages(fmt.Sprintf("SELECT * FROM images WHERE instrument IN ('CHEMCAM_RMI') ORDER BY utc DESC LIMIT %d", limit), db)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
		return
	}

	data.Latest = latest
	data.Mahli = mahli
	data.Mastcam = mastcam
	data.Navcam = navcam
	data.Hazcam = hazcam
	data.Chemcam = chemcam

	t, err := template.New("index.html").Funcs(template.FuncMap{
		"fromNow": fromNow,
	}).ParseFiles("index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), 500)
	}
}

func queryImages(query string, db *sql.DB) (marsimages.ListOfImages, error) {
	var loi marsimages.ListOfImages
	var marsImage marsimages.MarsImage

	rows, err := db.Query(query)
	if err != nil {
		return loi, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&marsImage.Instrument, &marsImage.ItemName, &marsImage.LMST, &marsImage.Sol, &marsImage.URL, &marsImage.UTC); err != nil {
			return loi, err
		}
		loi.Images = append(loi.Images, marsImage)
	}

	return loi, nil
}

func fromNow(utc string) string {
	t, err := parseDateTime(utc)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	dur := math.Floor(time.Since(t).Hours())
	// return string(t.Format("2006-01-02 15:04:05"))
	return fmt.Sprintf("%v hours ago", dur)
}

func parseDateTime(utc string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", utc)
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		golog.Panicf("%s environment variable not set.", k)
	}
	return v
}
