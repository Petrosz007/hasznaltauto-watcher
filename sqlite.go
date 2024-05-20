package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func Migrate(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStmt := `
  CREATE TABLE IF NOT EXISTS scans (
    time INTEGER NOT NULL, -- unix timestamp
    url  TEXT NOT NULL,
    UNIQUE(time, url)
  );
	CREATE TABLE IF NOT EXISTS listings (
	  id           TEXT NOT NULL PRIMARY KEY,
	  url          TEXT NOT NULL,
    first_seen   INTEGER NOT NULL,
    last_seen    INTEGER NOT NULL,
    scan_url     TEXT NOT NULL,
	  FOREIGN KEY(first_seen, scan_url) REFERENCES scans(time, url),
	  FOREIGN KEY(last_seen, scan_url) REFERENCES scans(time, url)
	);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return err
	}

	return nil
}

func writeNewScanTime(db *sql.DB, scanTime int64, scanUrl string) error {
	stmt, err := db.Prepare("INSERT INTO scans(time, url) values(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(scanTime, scanUrl)
	if err != nil {
		return err
	}

	return nil
}

func writeListings(db *sql.DB, scanTime int64, scanUrl string, listings []Listing) error {
	log.Printf("Writing %d listings into DB\n", len(listings))
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO listings(id, url, first_seen, last_seen, scan_url) values(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, listing := range listings {
		_, err := stmt.Exec(listing.ListingId, listing.Url, scanTime, scanTime, scanUrl)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func updateLastSeen(db *sql.DB, scanTime int64, listings []Listing) error {
	log.Printf("Updating %d listings in the DB\n", len(listings))
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
  UPDATE listings
  SET last_seen = ?
  WHERE listings.id = ?
  `)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for _, listing := range listings {
		_, err = stmt.Exec(scanTime, listing.ListingId)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func WriteScanToDb(dbPath string, scanTime int64, scanUrl string, listings []Listing) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	err = writeNewScanTime(db, scanTime, scanUrl)
	if err != nil {
		return err
	}

	err = writeListings(db, scanTime, scanUrl, listings)
	if err != nil {
		return err
	}

	err = updateLastSeen(db, scanTime, listings)
	if err != nil {
		return err
	}

	return nil
}

func readUrls(rows *sql.Rows) ([]string, error) {
	urls := make([]string, 0, 10)
	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		if err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	return urls, nil
}

func GetFirstSeenListingURLs(dbPath string, scanTime int64, scanUrl string) ([]string, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stmt, err := db.Prepare(`
  SELECT url FROM listings
  WHERE first_seen = ? AND scan_url = ?
  AND (SELECT COUNT(*) FROM scans WHERE url = ?) > 1
  `)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(scanTime, scanUrl, scanUrl)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return readUrls(rows)
}
