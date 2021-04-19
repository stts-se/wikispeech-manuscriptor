package main

import (
	"fmt"
	"os"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "USAGE: <Sqlite3 DB file> <FEATCAT-FILES>\nTab separated files of source-featname, target-featname, featvalue\n")
		os.Exit(0)
	}

	err := dbapi.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open db file '%s' : %v\n", os.Args[1], err)
		os.Exit(1)
	}

	for _, fn := range os.Args[2:] {

		sourceFeatName, cats, err := dbapi.ParseChunkFeatCatFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read featcat file '%s' : %v\n", os.Args[1], err)
			os.Exit(1)
		}

		n, err := dbapi.AddChunkFeatCats(sourceFeatName, cats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to add feat cats : %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Completed %s (inserted %d)\n", fn, n)
	}
}
