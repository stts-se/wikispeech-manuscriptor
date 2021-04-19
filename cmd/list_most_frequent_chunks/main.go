package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/stts-se/manuscriptor2000/dbapi"
)

func main() {

	var n = 40

	if len(os.Args) < 2 {

		fmt.Fprintf(os.Stdout, "USAGE: <SQLITE DB PATH> <NUMBER OF CHUNKS>? (defalt "+strconv.Itoa(n)+")")
		os.Exit(0)
	}

	if len(os.Args) > 2 {
		n0, err := strconv.ParseInt(os.Args[2], 10, 64)
		n = int(n0)

		if err != nil {
			fmt.Fprintf(os.Stderr, "failes to parse command line int arg : %v\n", err)
			os.Exit(1)
		}
	}

	dbapi.Open(os.Args[1])

	chunks, err := dbapi.SelectMostFrequentChunks(n)

	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to run DB query : %v\n", err)
		os.Exit(1)
	}

	for _, c := range chunks {
		fmt.Printf("%d:%d\t%s\n", c.Freq, c.ID, c.Text)
	}

}
