package main

import (
	"compress/bzip2"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
	"github.com/stts-se/wikispeech-manuscriptor/text"
)

// TODO decide on best values below
const (
	// minimal number of paragraphs in an article
	minParagraphs int = 4

	// minimal number of sentences in a paragraph
	minSentences int = 6

	// max number of tokens in a sentence
	maxTokens int = 70

	// min number of unique word forms in a sentence
	minUniqueWordForms int = 4
)

var (
	// accepted sentence end (sentences ending differently will be dropped)
	acceptedSentEndRE = regexp.MustCompile(`[.?!][”"]?$`)

	// sentences with double left curly brackets will be dropped
	doubleCurliesRE = regexp.MustCompile(`\{\{`)

	// TODO add regexp for strange/blockable tokens?
	//blockTokensRE = regexp.MustCompile(`TODOXXXXXXXXXXX`)
)

func keepArticle(a text.Article) bool {

	if len(a.Paragraphs) < minParagraphs {
		//fmt.Printf("TO FEW PARAS : %d %s\n", len(a.Paragraphs), a.URL)
		return false
	}

	n := 0
	for _, p := range a.Paragraphs {
		n += len(p.Sentences)
	}
	if n < minSentences {
		return false
	}

	return true
}

func removeUnwantedSents(a *text.Article) int {
	var res int
	for i, p := range a.Paragraphs {
		var sents []text.Sentence
		for _, s := range p.Sentences {
			if keepSent(s) {
				sents = append(sents, s)
			} else {

				//log.Printf("skipped %v\n", s)
				res++
			}
		}
		p.Sentences = sents
		a.Paragraphs[i] = p
	}

	return res
}

func keepSent(s text.Sentence) bool {

	if len(s.Feats[text.FeatWord])+len(s.Feats[text.FeatPunct]) > maxTokens {
		return false
	}

	if len(s.Feats[text.FeatWord]) < minUniqueWordForms {
		return false
	}

	if !acceptedSentEndRE.MatchString(s.Text) {
		return false
	}

	if doubleCurliesRE.MatchString(s.Text) {
		return false
	}

	// TODO add regexp blockTokensRE? (see above)
	// if blockTokensRE.MatchString(s.Text) {
	// 	return false
	// }

	return true
}

func bulkInsertChunkFeats(dbFile string, sents ...text.Sentence) {
	log.Printf("Bulk inserting chunkfeats...")
	nFeats, err := dbapi.BulkInsertChunkFeats(dbFile, sents...)
	if err != nil {
		log.Fatalf("Bulk insert failed: %v\n", err)
	}
	log.Printf("Inserted %d chunkfeats for %d sents", nFeats, len(sents))
}

// go run cmd/load_db/main.go <SQLITE3 DB FILE> <DUMP EXTRACTION DIR>/extracted/AA/wiki_00.bz2
func main() {

	bulkSize := flag.Int("bulk", 100000, "Bulk `size` for import (approx number of sentences)")
	help := flag.Bool("h", false, "Print usage and exit")

	flag.Parse()

	if len(flag.Args()) < 3 {
		fmt.Fprintf(os.Stderr, "USAGE: <options> <SQLITE3 DB FILE> <featcatdir> <(gzipped or plain) WikiExtractor.py output files>\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *help {
		fmt.Fprintf(os.Stderr, "USAGE: <options> <SQLITE3 DB FILE> <featcatdir> <gzipped WikiExtractor.py output files>\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	dbFile := flag.Args()[0]
	chunkFeatCatFolder := flag.Args()[1]

	fi, err := os.Stat(chunkFeatCatFolder)
	mode := fi.Mode()
	if !mode.IsDir() {
		log.Fatalf("Expected folder, found file: %s", chunkFeatCatFolder)
	}

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Fatalf("no such file %s\n", dbFile)
	}

	err = dbapi.Open(dbFile, "PRAGMA synchronous = OFF", "PRAGMA journal_mode = MEMORY", "PRAGMA foreign_keys=OFF", "PRAGMA cache_size=10000")
	if err != nil {

		log.Fatalf("failed to open db file '%s' : %v", dbFile, err)
	}

	aN := 0
	sN := 0
	parasSkipped := 0
	sentsSkipped := 0

	//log.Println("PROBLEMATIC: different runs of WikiExtrator.py appear to use different paragraph delimiter?")

	sents := []text.Sentence{}
	nArticles := 0

	for _, fn := range flag.Args()[2:] {
		//fmt.Println(fn)

		fh, err := os.Open(fn)
		if err != nil {
			log.Fatalf("Failed open file : %v", err)
		}

		var bts []byte
		if strings.HasSuffix(fn, ".bz2") {
			r := bzip2.NewReader(fh)
			bts, err = ioutil.ReadAll(r)
			if err != nil {
				log.Fatalf("Failed to read FileReader : %v", err)
			}

		} else {
			bts, err = ioutil.ReadAll(fh)
			if err != nil {
				log.Fatalf("Failed to read file '%s' : %v", fn, err)
			}

		}

		articles := text.ExtractedFile2Articles(string(bts))

		log.Printf("Read file '%s'\n", fn)

		for _, a := range articles {
			nArticles++

			//fmt.Printf("INCOMING:\n%v\n", a)

			sentsSkipped += removeUnwantedSents(&a)

			if !keepArticle(a) {
				//fmt.Printf("OUTGOING:\n%v\n\n", a)

				//log.Printf("skipped %v\n", a.URL)
				parasSkipped++
				continue
			}

			for _, p := range a.Paragraphs {
				sN += len(p.Sentences)
			}

			aN++
			_, insertedSents, err := dbapi.Add(a, false) // false = insert chunk feats in bulk later
			for _, s := range insertedSents {
				sents = append(sents, s)
			}
			if err != nil {
				log.Fatalf("Failed to add article : %v", err)
			}

			//if nArticles%*bulkSize == 0 {
			if len(sents) >= *bulkSize {
				bulkInsertChunkFeats(dbFile, sents...)
				sents = []text.Sentence{}
			}

		}
	}

	log.Printf("Articles skipped: %d\n", parasSkipped)
	log.Printf("Articles added: %d\n", aN)
	log.Printf("Sentenes skipped: %d\n", sentsSkipped)
	log.Printf("Sentences added: %d\n", sN)

	// bulk insert chunkfeats
	if len(sents) > 0 {
		bulkInsertChunkFeats(dbFile, sents...)
	}

	// generate word freqs
	log.Println("Generating word frequency table...")
	err = dbapi.PopulateWordFreqTable()
	if err != nil {
		log.Fatalf("failed to create word frequency table : %v", err)
	}
	log.Println("Done generating word frequency table!")

	// generate lowest word freqs
	log.Println("Generating lowest word frequency per chunk...")
	err = dbapi.InsertLowestWordFreqForChunk()
	if err != nil {
		log.Fatalf("failed to insert lowest word frequency per chunk : %v", err)
	}

	log.Println("Done generating lowest word frequency per chunk!")

	// load chunk feat cats
	chunkFeatCatFiles, _ := filepath.Glob(filepath.Join(chunkFeatCatFolder, "*.txt"))
	for _, fn := range chunkFeatCatFiles {
		sourceFeatName, cats, err := dbapi.ParseChunkFeatCatFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read featcat file '%s' : %v\n", fn, err)
			os.Exit(1)
		}

		n, err := dbapi.AddChunkFeatCats(sourceFeatName, cats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to add feat cats : %v\n", err)
			os.Exit(1)
		}
		log.Printf("Completed %s (inserted %d)\n", fn, n)
	}

}
