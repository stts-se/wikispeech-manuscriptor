package filter

import (
	//"fmt"

	"encoding/json"
	"log"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
	"github.com/stts-se/wikispeech-manuscriptor/protocol"
	"github.com/stts-se/wikispeech-manuscriptor/text"
)

var testDbPath = "tst_manuscript_filter.db"

func init() {

	log.Println("INITIALISING filter TESTS")
	err := os.RemoveAll(testDbPath)
	if err != nil {
		log.Printf("failed to remove test file '%s' : %v", testDbPath, err)
	}

	err = dbapi.CreateDB(testDbPath, path.Join("..", "dbapi", "schema_sqlite.sql"))
	if err != nil {
		log.Fatalf("failed to create db '%s' : %v", testDbPath, err)
	}
}

func TestFilter1(t *testing.T) {
	batchName := "test_batch_1"
	sents := []string{
		"Det bor en del stockholmare i området",
		"i många städer finns parker",
		"Kända personer från Lviv",
		"Stockholm ligger inte @ Skåne.",
		`Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}

	expectBatch := []string{
		"Det bor en del stockholmare i området",
		"i många städer finns parker",
		"Kända personer från Lviv",
	}

	textSents := []text.Sentence{}
	for _, s := range sents {
		textSents = append(textSents, text.ComputeSentence(s))
	}

	a := text.Article{
		URL: "testbatchmetadata:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: textSents},
		},
	}

	// add sents
	aID, _, err := dbapi.Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	// filter => batch
	filterConfig := protocol.FilterPayload{
		BatchName:  batchName,
		TargetSize: 100,
		Opts: []protocol.FilterOpt{
			{Name: WordCount, Args: []string{"4", "11"}},
			{Name: CommaCount, Args: []string{"0", "25"}},
			{Name: ExcludeChunkRE, Args: []string{`[\p{Greek}]`}},
			{Name: ExcludeChunkRE, Args: []string{`[^a-zA-ZåäöÅÄÖéÉüÜ0-9 ,$€£.!?/()"':—–-]`}},
		},
	}
	batchMetadata := protocol.BatchMetadata{FilterPayload: filterConfig}
	filterQueryBuilder, err := NewQueryBuilder(filterConfig)
	if err != nil {
		t.Errorf("Couldn't create query builder : %v", err)
		return
	}

	n, err := ExecQuery(filterQueryBuilder)
	if err != nil {
		t.Errorf("Couldn't exec query : %v", err)
	}

	batchMetadata.OutputSize = int(n)

	pBytes, err := json.Marshal(batchMetadata)
	if err != nil {
		t.Errorf("failed to marshal filter metadata : %v", err)
		return
	}
	err = dbapi.SetBatchProperties(filterConfig.BatchName, pBytes)
	if err != nil {
		t.Errorf("failed to save batch properties : %v", err)
		return
	}

	rows, err := dbapi.ExecQuery("SELECT chunk.text FROM chunk, batch WHERE batch.name = ? AND chunk.id = batch.chunk_id", []interface{}{batchName})
	if err != nil {
		t.Errorf("failed to read batches : %v", err)
	}
	gotSents := []string{}
	for rows.Next() {
		var name string
		rows.Scan(&name)
		gotSents = append(gotSents, name)
	}
	if !reflect.DeepEqual(expectBatch, gotSents) {
		t.Errorf("Expected %v, got %v", expectBatch, gotSents)
	}

}
