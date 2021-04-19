package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pkg/profile"

	"github.com/stts-se/manuscriptor2000/dbapi"
	"github.com/stts-se/manuscriptor2000/text"
)

func main() {

	profiling := flag.Bool("profile", false, "Profiling")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "USAGE: <Sqlite3 db file>\n")
		os.Exit(0)
	}

	fn := flag.Args()[0]
	err := dbapi.Open(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db '%s' : %v\n", fn, err)
		os.Exit(1)
	}

	tx, err := dbapi.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start db transaction : %v\n", err)
		os.Exit(1)
	}

	if *profiling {
		// "github.com/pkg/profile"
		defer profile.Start().Stop()
	}

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't commit db transaction : %v\n", err)
		os.Exit(1)
	}
	tx, err = dbapi.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start db transaction : %v\n", err)
		os.Exit(1)
	}
	// err = dbapi.ForeignKeysOff()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Failed ForeignKeysOff : %v\n", err)
	// 	os.Exit(1)
	// }

	err = dbapi.SynchronousOff()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed SynchronousOff : %v\n", err)
		os.Exit(1)
	}

	err = dbapi.CacheSize(10000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed CacheSize : %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Reading sents from db... ")

	// Go through every single chunk...
	rows, err := tx.Query(`SELECT id, text FROM chunk`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed select from table chunk : %v\n", err)
		tx.Rollback()
		os.Exit(1)
	}

	sents := []text.Sentence{}
	for rows.Next() {
		var id int64
		var txt string
		err = rows.Scan(&id, &txt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scan from table chunk : %v\n", err)
			tx.Rollback()
			os.Exit(1)
		}
		if len(sents)%1000 == 0 {
			fmt.Fprintf(os.Stderr, "\rReading sents from db... %d", len(sents))
		}

		sent := text.ComputeSentence(txt)
		sent.ID = id
		sents = append(sents, sent)
	}
	fmt.Fprintf(os.Stderr, "\rReading sents from db... done (%d sents)\n", len(sents))

	err = tx.Commit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't commit transaction : %v\n", err)
		os.Exit(1)
	}
	// err = dbapi.Close()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Couldn't close db : %v\n", err)
	// 	os.Exit(1)
	// }

	nFeats, err := dbapi.BulkInsertChunkFeats(fn, sents...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bulk insert failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Inserted %d featcats for %d sents\n", nFeats, len(sents))

}
