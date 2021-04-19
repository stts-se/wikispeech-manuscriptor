package filter

import (
	//"fmt"
	"reflect"
	"testing"

	"github.com/stts-se/manuscriptor2000/protocol"
)

func TestPayloadOpt2FilterOpt(t *testing.T) {
	var input protocol.FilterOpt
	var err error
	var got func(*queryBuilder)
	var qb = &queryBuilder{}
	var expectArgs []interface{}

	// paragraph count
	qb = &queryBuilder{}
	input = protocol.FilterOpt{Name: ParagraphCount, Args: []string{"10", "20"}}
	got, err = payloadOpt2filterOpt(input)
	if err != nil {
		t.Errorf("Couldn't parse payload opt %v: %v", input, err)
	}
	expectArgs = []interface{}{"paragraph_count", 10, 20}

	got(qb)
	if !reflect.DeepEqual((*qb).args, expectArgs) {
		t.Errorf("Expected %v, found %v", expectArgs, (*qb).args)
	}

	// sentence count
	qb = &queryBuilder{}
	input = protocol.FilterOpt{Name: SentenceCount, Args: []string{"10", "-1"}}
	got, err = payloadOpt2filterOpt(input)
	if err != nil {
		t.Errorf("Couldn't parse payload opt %v: %v", input, err)
	}
	expectArgs = []interface{}{"sentence_count", 10}

	got(qb)
	if !reflect.DeepEqual((*qb).args, expectArgs) {
		t.Errorf("Expected %v, found %v", expectArgs, (*qb).args)
	}

	// sentence count
	qb = &queryBuilder{}
	input = protocol.FilterOpt{Name: SentenceCount, Args: []string{"-1", "20"}}
	got, err = payloadOpt2filterOpt(input)
	if err != nil {
		t.Errorf("Couldn't parse payload opt %v: %v", input, err)
	}
	expectArgs = []interface{}{"sentence_count", 20}

	got(qb)
	if !reflect.DeepEqual((*qb).args, expectArgs) {
		t.Errorf("Expected %v, found %v", expectArgs, (*qb).args)
	}

}

func TestQueryBuilderFromPayload(t *testing.T) {

	filterConfig := protocol.FilterPayload{
		BatchName:  "test_batch_1",
		TargetSize: 400000,
		Opts: []protocol.FilterOpt{
			{Name: WordCount, Args: []string{"4", "25"}},
			{Name: CommaCount, Args: []string{"0", "25"}},
			{Name: SourceRE, Args: []string{"00$"}},
			{Name: ParagraphCount, Args: []string{"10", "20"}},
			{Name: SentenceCount, Args: []string{"10", "-1"}},
			{Name: DigitCount, Args: []string{"0"}},
			{Name: ChunkFeatCats, Args: []string{"se_place", "se_name", "se_calendar"}},
			{Name: LowestWordFreq, Args: []string{"2"}},
			{Name: ExcludeChunkRE, Args: []string{`[\p{Greek}]`}},
			{Name: ExcludeChunkRE, Args: []string{`[^a-zA-ZåäöÅÄÖéÉüÜ0-9 ,$€£@.!?/()"':—–-]`}},
			{Name: ExcludeBatches, Args: []string{"test1", "test2"}},
		},
	}
	_, err := NewQueryBuilder(filterConfig)
	if err != nil {
		t.Errorf("Got error from NewQueryBuilder: %v", err)
	}
}
