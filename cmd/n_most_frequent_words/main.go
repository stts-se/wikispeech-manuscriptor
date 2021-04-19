package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
)

func main() {

	var n = 40

	if len(os.Args) < 2 {

		fmt.Fprintf(os.Stdout, "USAGE: <SQLITE DB PATH> <NUMBER MOST FREQUENT WORDS>? (defalt "+strconv.Itoa(n)+")")
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

	words, err := dbapi.MostFrequentWords(n)

	if err != nil {

		fmt.Fprintf(os.Stderr, "failed to run DB query : %v\n", err)
		os.Exit(1)
	}

	for _, w := range words {
		fmt.Printf("%s\n", w)
	}

}
