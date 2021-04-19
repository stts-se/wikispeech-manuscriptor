package filter

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stts-se/manuscriptor2000/text"
)

func f() { fmt.Println() }

func TestQueryBuilder(t *testing.T) {

	batchName := "new_batch_01"

	//
	qb1, err := newFilterQueryBuilder(
		filterQHeadInto(batchName),
		wordCountView(4, 19),
		commaCountView(0, 3),
		nDigitCountView(0),
		lowestWordFreqCountView(500),
		//tailNotInBatchOrderByLowestWordFreq(batchName, 10000),
	)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}

	if qb1.head == "" {
		t.Errorf("wanted string, got empty string")
	}

	if w, g := 7, len(qb1.args); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	//
	qb2, err := newFilterQueryBuilder(
		filterQHeadInto(batchName),
		wordCountView(4, 25),
		commaCountView(-1, 4),
		nDigitCountView(0),
		chunkFeatCat(text.FeatSEPlace),
		//excludePuncts(";", "!"),
		excludeChunkRE("ö"),
		//lowestWordFreqCountView(15),
		tailNotInBatches(batchName),
		tailLimit(100000),
		//tailNotInBatchOrderByLowestWordFreq(excludeBatch, n),
	)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}

	s, a := qb2.query()
	fmt.Printf("%#v\n%#v\n", s, a)

}

var whereRE = regexp.MustCompile("(?i)WHERE +chunk.id.*WHERE +chunk.id")

func TestMinimalQueryBuilder(t *testing.T) {

	batchName := "new_batch_01"

	//
	qb3, err := newFilterQueryBuilder(
		filterQHeadInto(batchName),
		tailNotInBatches("nizze"),
		tailLimit(100000),
	)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}

	q, _ := qb3.query()

	if whereRE.MatchString(q) {
		t.Errorf("Output query contains two WHERE conditions: %s", q)
	}

}
