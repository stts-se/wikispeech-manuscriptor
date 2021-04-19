package main

import (
	"compress/bzip2"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/stts-se/wikispeech-manuscriptor/text"
)

// go run cmd/extracted_article_filter/main.go <DUMP EXTRACTION DIR>/extracted/AA/wiki_00.bz2
func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: <gzipped WikiExtractor.py output files>\n")
	}

	for _, fn := range os.Args[1:] {
		//fmt.Println(fn)

		fh, err := os.Open(fn)
		if err != nil {
			log.Fatalf("Failed open file : %v", err)
		}

		r := bzip2.NewReader(fh)
		bts, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatalf("Failed to read FileReader : %v", err)
		}

		articles := text.ExtractedFile2Articles(string(bts))

		for _, a := range articles {
			fmt.Printf("%#v\n", a)
			fmt.Println("===============")
		}

	}

}
