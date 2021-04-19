package dbapi

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"time"

	//"database/sql"
	"log"
	"os"
	"testing"

	"github.com/stts-se/wikispeech-manuscriptor/protocol"
	"github.com/stts-se/wikispeech-manuscriptor/text"
)

var testDbPath = "tst_manuscript_dbapi.db"

func init() {

	log.Println("INITIALISING dbapi TESTS")
	err := os.RemoveAll(testDbPath)
	if err != nil {
		log.Printf("failed to remove test file '%s' : %v", testDbPath, err)
	}

	err = CreateDB(testDbPath, "schema_sqlite.sql")
	if err != nil {
		log.Fatalf("failed to create db '%s' : %v", testDbPath, err)
	}

	// err = Open(testDbPath)
	// if err != nil {
	// 	log.Fatalf("failed to open test db file '%s' : %v", testDbPath, err)
	// }

	// err = ExecSchema("schema_sqlite.sql")
	// if err != nil {
	// 	log.Fatalf("failed to ExecSchema : %v", err)
	// }
	//Open(f)

}

// TODO Look up inserted values
func TestInsertSource(t *testing.T) {

	id1, err := InsertSource("s1")
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := int64(1), id1; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	id1, err = InsertSource("s1")
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := int64(1), id1; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	id2, err := InsertSource("s2")
	if err != nil {
		t.Errorf("%v", err)
	}
	if w, g := int64(2), id2; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

}

// TODO Look up inserted values
func TestInsertChunk(t *testing.T) {
	id1, id2, err := InsertChunk("s1", "I am a tea pot.")
	if err != nil {
		t.Errorf("failure : %v", err)
	}

	if w, g := int64(1), id1; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := int64(1), id2; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	id3, id4, err := InsertChunk("s1", "I am a coffe pot.")
	if err != nil {
		t.Errorf("failure : %v", err)
	}

	if w, g := int64(1), id3; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := int64(2), id4; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

}

func TestInsertChunkFeats1(t *testing.T) {

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		t.Errorf("InsertChunkFeats failed to begin transaction : %v", err)
	}

	sourceID, chunkID, _, err := InsertChunkTx(tx, "sourcex", "I am a happy-sad cow")
	if err != nil {
		t.Errorf("failed InsertChunk : %v", err)
		return
	}

	if sourceID < 1 {
		t.Errorf("got zero ID")
		return
	}

	if chunkID < 1 {
		t.Errorf("got zero ID")
		return
	}

	feats := map[string]map[string]int{
		"token_num": {"11": 1},
	}

	err = InsertChunkFeatsTx(tx, chunkID, feats)
	if err != nil {
		tx.Rollback()
		t.Errorf("%v", err)
		return
	}

	feats = map[string]map[string]int{
		"SPECIAL_FEAT": {"special_value": 11},
	}

	err = InsertChunkFeatsTx(tx, chunkID, feats)
	if err != nil {
		tx.Rollback()
		t.Errorf("%v", err)
		return
	}

	sents, err := GetSentsTx(tx, chunkID)
	if len(sents) != 1 {
		tx.Rollback()
		t.Errorf("Expected %d sents, found %d", 1, len(sents))
		return
	}
	expectFeats := map[string]map[string]int{
		"token_num":    {"11": 1},
		"SPECIAL_FEAT": {"special_value": 11},
	}
	sent := sents[0]
	if !reflect.DeepEqual(sent.Feats, expectFeats) {
		t.Errorf("Expected %#v, got %#v\n", expectFeats, sent.Feats)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("commit failed : %v", err)
	}
}

func TestInsertChunkFeats2(t *testing.T) {
	c := "I'm a lonely stupid brick."
	feats := map[string]map[string]int{
		"f1": {"v1": 13},
		"f2": {"v2": 14},
		"f3": {"v1": 15},
		"f4": {"v2": 16},
		"f5": {"v1": 17},
	}

	_, cID, err := InsertChunk("sourceX", c)

	if err != nil {
		t.Errorf("Bomb! : %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		t.Errorf("InsertChunkFeats failed to begin transaction : %v", err)
	}

	err = InsertChunkFeatsTx(tx, cID, feats)
	if err != nil {
		tx.Rollback()
		t.Errorf("Really...? : %v", err)
		return
	}

	sents, err := GetSentsTx(tx, cID)
	if len(sents) != 1 {
		tx.Rollback()
		t.Errorf("Expected %d sents, found %d", 1, len(sents))
		return
	}
	sent := sents[0]
	if !reflect.DeepEqual(sent.Feats, feats) {
		t.Errorf("Expected %#v, got %#v\n", feats, sent.Feats)
	}
	err = tx.Commit()
	if err != nil {
		t.Errorf("commit failed : %v", err)
	}

}

// TODO Look up inserted values
func TestInsertSourceFeats(t *testing.T) {

	s := "article 1a"
	sID, err := InsertSource(s)
	if err != nil {
		t.Errorf("InsertSource : %v", err)
	}

	fs := []Feat{
		{Name: "sf1", Value: "sv1"},
		{Name: "sf2", Value: "sv1"},
		{Name: "sf3", Value: "sv2"},
		{Name: "sf4", Value: "sv2"},
	}

	tx, err := Begin()
	if err != nil {
		t.Errorf("failed to begin transaction : %v", err)
	}

	err = InsertSourceFeatsTx(tx, sID, fs)
	if err != nil {
		t.Errorf("failed InsertSourceFeatsTx : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		t.Errorf("commit failed : %v", err)
	}

}

func TestBlockSent(t *testing.T) {
	blockableSent1 := `Detta är en. tras-x mening.`
	blockableSent2 := `€#¤!.`
	blockableSents := []string{blockableSent1, blockableSent2}

	a := text.Article{
		URL: "testblocksent:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: []text.Sentence{
				text.ComputeSentence(`Hej din gamla apa!`),
				text.ComputeSentence(blockableSent1),
				text.ComputeSentence(`Quack?`),
			}},
			{Sentences: []text.Sentence{
				text.ComputeSentence(`Jag springer till skogs!`),
				text.ComputeSentence(`Vintern kommer snart, det var som 17.`),
				text.ComputeSentence(blockableSent2),
			}},
		},
	}

	aID, _, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	err = BlockSents(blockableSents...)
	if err != nil {
		t.Errorf("BlockSent failed : %v", err)
		return
	}

	tx, err := Begin()
	if err != nil {
		log.Fatalf("Begin failed : %v", err)
		return
	}
	rows, err := tx.Query("SELECT chunk.text FROM batch, chunk WHERE chunk.id = batch.chunk_id AND batch.name = ?", text.BlockBatch)
	if err != nil {
		tx.Rollback()
		t.Errorf("Query failed : %v", err)
		return
	}
	res := []string{}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			tx.Rollback()
			t.Errorf("failed to scan rows : %v", err)
			return
		}
		res = append(res, s)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		t.Errorf("error when reading result row : %v", err)
		return
	}
	rows.Close() // strictly not needed
	err = tx.Commit()
	if err != nil {
		t.Errorf("Commit failed : %v", err)
		return
	}

	if !reflect.DeepEqual(blockableSents, res) {
		t.Errorf("expected %v, found %v", blockableSents, res)
		return
	}
}

func TestAdd(t *testing.T) {

	a := text.Article{
		URL: "testadd:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: []text.Sentence{
				text.ComputeSentence(`Let's boogaloo!`),
				text.ComputeSentence(`Don't be sad, this is soon over.`),
				text.ComputeSentence(`Quack?`),
			}},
			{Sentences: []text.Sentence{
				text.ComputeSentence(`I hate you!`),
				text.ComputeSentence(`No, you don't.`),
				text.ComputeSentence(`Flimflam?`),
			}},
		},
	}

	aID, _, err := Add(a, true)
	if err != nil {
		t.Errorf("This didn't go well : %v", err)
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
	}
}

func TestAddDupes(t *testing.T) {

	a := text.Article{
		URL: "testadd:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: []text.Sentence{
				text.ComputeSentence(`x Let's boogaloo!`),
				text.ComputeSentence(`x Don't be sad, this is soon over.`),
				text.ComputeSentence(`x Don't be sad, this is soon over.`),
				text.ComputeSentence(`x Quack?`),
			}},
			{Sentences: []text.Sentence{
				text.ComputeSentence(`x I hate you!`),
				text.ComputeSentence(`x Quack?`),
				text.ComputeSentence(`x No, you don't.`),
				text.ComputeSentence(`x Flimflam?`),
			}},
		},
	}

	aID, inserted, err := Add(a, true)
	if err != nil {
		t.Errorf("This didn't go well : %v", err)
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
	}
	if len(inserted) != 6 {
		t.Errorf("Expected %d inserted sents, found %d", 6, len(inserted))
	}
}

func TestAddChunkFeatCats(t *testing.T) {
	sentsWithFeats := map[string]map[string]int{
		"A Det bor en del stockholmare i området": {},
		"A i många städer finns parker":           {},
		"A Kända personer från Lviv":              {"lviv": 1},
		"A Stockholm ligger inte i Skåne.":        {"skåne": 1, "stockholm": 1},
		`A Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`: {"amsterdam": 2},
		"A Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848":                                    {"wisconsin": 1, "usa": 1},
	}
	sents := []string{}
	for s := range sentsWithFeats {
		sents = append(sents, s)
	}
	textSents := []text.Sentence{}
	for _, s := range sents {
		textSents = append(textSents, text.ComputeSentence(s))
	}

	a := text.Article{
		URL: "testblocksent:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: textSents},
		},
	}

	// add sents
	aID, _, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	// add chunkfeat cats
	places := []ChunkFeatCat{
		{TargetFeatName: "place_on_earth", FeatValue: "stockholm"},
		{TargetFeatName: "place_on_earth", FeatValue: "skåne"},
		{TargetFeatName: "place_on_earth", FeatValue: "lviv"},
		{TargetFeatName: "place_on_earth", FeatValue: "wisconsin"},
		{TargetFeatName: "place_on_earth", FeatValue: "amsterdam"},
		{TargetFeatName: "place_on_earth", FeatValue: "usa"},
	}
	n, err := AddChunkFeatCats("word", places)
	if n != len(sentsWithFeats) {
		t.Errorf("Expected %d, got %d", len(sentsWithFeats), n)
		return
	}

	// the tests
	tx, err := Begin()
	if err != nil {
		log.Fatalf("Begin failed : %v", err)
		return
	}
	args := []interface{}{}
	for _, sent := range sents {
		args = append(args, sent)
	}
	rows, err := tx.Query("SELECT chunk.id FROM chunk WHERE chunk.text IN (?,?,?,?,?,?)", args...)
	if err != nil {
		tx.Rollback()
		t.Errorf("Query failed : %v", err)
		return
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			tx.Rollback()
			t.Errorf("failed to scan rows : %v", err)
			return
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		t.Errorf("error when reading result row : %v", err)
		return
	}
	rows.Close() // strictly not needed
	err = tx.Commit()
	if err != nil {
		t.Errorf("Commit failed : %v", err)
		return
	}

	res, err := GetSents(ids...)
	sort.Slice(res, func(i, j int) bool { return strings.ToLower(res[i].Text) < strings.ToLower(res[j].Text) })
	if len(res) != len(sents) {
		t.Errorf("Expected %d sents, found %d", len(sents), len(res))
	}
	for _, s := range res {
		expFeats := sentsWithFeats[s.Text]
		sFeats := s.Feats["place_on_earth"]
		if len(expFeats) == 0 {
			if len(sFeats) != 0 {
				t.Errorf("Expected %v, found %v", expFeats, sFeats)
			}
		} else if !reflect.DeepEqual(expFeats, sFeats) {
			t.Errorf("Expected %v, found %v", expFeats, sFeats)
		}
	}

}

func TestBatchMetadata(t *testing.T) {
	batchName := "test_batch_1"
	sents := []string{
		"B Det bor en del stockholmare i området",
		"B i många städer finns parker",
		"B Kända personer från Lviv",
		"B Stockholm ligger inte @ Skåne.",
		`B Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"B Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
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
	aID, sentsWithID, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	// filter to batch
	filterConfig := protocol.FilterPayload{
		BatchName:  batchName,
		TargetSize: 100,
		Opts: []protocol.FilterOpt{
			{Name: "word_count", Args: []string{"4", "25"}},
			{Name: "comma_count", Args: []string{"0", "25"}},
			{Name: "exclude_chunk_re", Args: []string{`[\p{Greek}]`}},
			{Name: "exclude_chunk_re", Args: []string{`[^a-zA-ZåäöÅÄÖéÉüÜ0-9 ,$€£.!?/()"':—–-]`}},
		},
	}
	inputMetadata := protocol.BatchMetadata{FilterPayload: filterConfig}
	inputMetadata.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	addToBatch := map[string]bool{
		"B Det bor en del stockholmare i området": true,
		"B i många städer finns parker":           true,
		"B Kända personer från Lviv":              true,
	}

	insert := "INSERT INTO batch (chunk_id, name) VALUES (?,?)"
	n := 0
	for _, s := range sentsWithID {
		if _, doAdd := addToBatch[s.Text]; doAdd {
			_, err = db.Exec(insert, s.ID, batchName)
			if err != nil {
				t.Errorf("add to batch failed : %v", err)
				return
			}
			n++
		}
	}
	if n != len(addToBatch) {
		t.Errorf("expected %d, got %d", len(addToBatch), n)
		return
	}

	inputMetadata.OutputSize = n

	pBytes, err := json.Marshal(inputMetadata)
	if err != nil {
		t.Errorf("failed to marshal filter metadata : %v", err)
		return
	}
	err = SetBatchProperties(filterConfig.BatchName, pBytes)
	if err != nil {
		t.Errorf("failed to save batch properties : %v", err)
		return
	}

	outputMetadata, err := GetBatchProperties(batchName)
	if !reflect.DeepEqual(inputMetadata, outputMetadata) {
		t.Errorf("expected %v, got %v", inputMetadata, outputMetadata)
		return
	}

}

func TestGetSetScriptProperties(t *testing.T) {
	scriptName := "test_script_1"
	selectorPayload := protocol.SelectorPayload{
		Options: protocol.SelectorOptions{
			FeatureOpts: []protocol.SelectorFeatOpt{
				{text.FeatBigramTransition, 0},
				{text.FeatBigram + "_top800", 3},
				{text.FeatFinalTrigram, 0},
				{text.FeatInitialBigram, 0},
				{text.FeatWord, 0},
				{text.FeatBigram, 0},
			},
			AdjustScoreForSentenceLength: false,
			AccumulatedScripts:           []string{},
			TargetSize:                   3,
			ScriptName:                   scriptName,
			FromBatch:                    "no_batch",
			//ContinuousPrint:              false,
			PrintMetaData: true,
			Debug:         false,
			Mode:          "rand",

			// rand selector
			CutoffRand:        100,
			MinIterationsRand: 50,

			// exhaustive selector
			ChunkSize:     0,
			ChunkDecrease: 0,
		},
	}
	inputMetadata := protocol.ScriptMetadata{SelectorPayload: selectorPayload}
	inputMetadata.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	inputMetadata.OutputSize = 7

	pBytes, err := json.Marshal(inputMetadata)
	if err != nil {
		t.Errorf("failed to marshal filter metadata : %v", err)
		return
	}
	err = SetScriptProperties(selectorPayload.Options.ScriptName, pBytes)
	if err != nil {
		t.Errorf("failed to save batch properties : %v", err)
		return
	}

	outputMetadata, err := GetScriptProperties(scriptName)
	if !reflect.DeepEqual(inputMetadata, outputMetadata) {
		t.Errorf("expected\n%#v, got\n%#v", inputMetadata, outputMetadata)
		return
	}

}

func TestGetScriptWithMetadata(t *testing.T) {
	scriptName := "test_script_2"
	sents := []string{
		"C Det bor en del stockholmare i området",
		"C i många städer finns parker",
		"C Kända personer från Lviv",
		"C Stockholm ligger inte @ Skåne.",
		`C Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"C Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}

	textSents := []text.Sentence{}
	for _, s := range sents {
		textSents = append(textSents, text.ComputeSentence(s))
	}

	a := text.Article{
		URL: "testgetscript:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: textSents},
		},
	}

	// add sents
	aID, sentsWithID, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	selectorPayload := protocol.SelectorPayload{
		Options: protocol.SelectorOptions{
			FeatureOpts: []protocol.SelectorFeatOpt{
				{text.FeatBigramTransition, 0},
				{text.FeatBigram + "_top800", 3},
				{text.FeatFinalTrigram, 0},
				{text.FeatInitialBigram, 0},
				{text.FeatWord, 0},
				{text.FeatBigram, 0},
			},
			AdjustScoreForSentenceLength: false,
			AccumulatedScripts:           []string{},
			TargetSize:                   3,
			ScriptName:                   scriptName,
			FromBatch:                    "no_batch",
			//ContinuousPrint:              false,
			PrintMetaData: true,
			Debug:         false,
			Mode:          "rand",

			// rand selector
			CutoffRand:        100,
			MinIterationsRand: 50,

			// exhaustive selector
			ChunkSize:     0,
			ChunkDecrease: 0,
		},
	}
	inputMetadata := protocol.ScriptMetadata{SelectorPayload: selectorPayload}
	addToScript := map[string]bool{
		"C Det bor en del stockholmare i området": true,
		"C i många städer finns parker":           true,
		"C Kända personer från Lviv":              true,
	}

	insert := "INSERT INTO script (chunk_id, name) VALUES (?,?)"
	n := 0
	for _, s := range sentsWithID {
		if _, doAdd := addToScript[s.Text]; doAdd {
			_, err = db.Exec(insert, s.ID, scriptName)
			if err != nil {
				t.Errorf("add to script failed : %v", err)
				return
			}
			n++
		}
	}
	if n != len(addToScript) {
		t.Errorf("expected %d, got %d", len(addToScript), n)
		return
	}

	inputMetadata.OutputSize = n

	pBytes, err := json.Marshal(inputMetadata)
	if err != nil {
		t.Errorf("failed to marshal filter metadata : %v", err)
		return
	}
	err = SetScriptProperties(selectorPayload.Options.ScriptName, pBytes)
	if err != nil {
		t.Errorf("failed to save script properties : %v", err)
		return
	}

	outputMetadata, err := GetScriptProperties(scriptName)
	if !reflect.DeepEqual(inputMetadata, outputMetadata) {
		t.Errorf("expected %v, got %v", inputMetadata, outputMetadata)
		return
	}

	script, err := GetScript(scriptName, 0, 0)

	if len(script) != len(addToScript) {
		t.Errorf("Expected %d, got %d", len(addToScript), len(script))
		return
	}

	for _, s := range script {
		if _, shouldBeAdded := addToScript[s.Text]; !shouldBeAdded {
			t.Errorf("Found %s in script, expected %v", s.Text, addToScript)
		}
	}

}

func contains(slice []string, s string) bool {
	for _, s0 := range slice {
		if s0 == s {
			return true
		}
	}
	return false
}

func TestGetScriptWithPages(t *testing.T) {
	scriptName := "test_script_3"
	sents := []string{
		"D Det bor en del stockholmare i området",
		"D i många städer finns parker",
		"D Kända personer från Lviv",
		"D Jag heter inte Kalle",
		"D Stockholm ligger inte @ Skåne.",
		`D Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"D Idag kommer nya restriktioner",
		"D Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}

	addToScript := []string{
		"D Det bor en del stockholmare i området",
		"D i många städer finns parker",
		"D Kända personer från Lviv",
		"D Stockholm ligger inte @ Skåne.",
		`D Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"D Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}

	textSents := []text.Sentence{}
	for _, s := range sents {
		textSents = append(textSents, text.ComputeSentence(s))
	}

	a := text.Article{
		URL: "testgetscript:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: textSents},
		},
	}

	// add sents
	aID, sentsWithID, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	insert := "INSERT INTO script (chunk_id, name) VALUES (?,?)"
	n := 0
	for _, s := range sentsWithID {
		if contains(addToScript, s.Text) {
			_, err = db.Exec(insert, s.ID, scriptName)
			if err != nil {
				t.Errorf("add to script failed : %v", err)
				return
			}
			n++
		}
	}
	if n != len(addToScript) {
		t.Errorf("expected %d, got %d", len(addToScript), n)
		return
	}

	// test pages
	var expect, got []string
	var script []text.Sentence

	var script2sents = func(script []text.Sentence) []string {
		res := []string{}
		for _, s := range script {
			res = append(res, s.Text)
		}
		return res
	}

	//
	expect = addToScript
	script, err = GetScript(scriptName, 0, 0)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 6 {
		t.Errorf("Expected 6 sents in script, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	//
	expect = addToScript[0:5]
	script, err = GetScript(scriptName, 1, 5)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 5 {
		t.Errorf("Expected 5 sents in script, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	//
	expect = []string{addToScript[5]}
	script, err = GetScript(scriptName, 2, 5)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 1 {
		t.Errorf("Expected 1 sent in script, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}
}

func TestGetBatchWithPages(t *testing.T) {
	batchName := "test_batch_3"
	sents := []string{
		"Det bor en del stockholmare i området",
		"i många städer finns parker",
		"Kända personer från Lviv",
		"Jag heter inte Kalle",
		"Stockholm ligger inte @ Skåne.",
		`Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"Idag kommer nya restriktioner",
		"Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}
	addToBatch := []string{
		"Det bor en del stockholmare i området",
		"i många städer finns parker",
		"Kända personer från Lviv",
		"Stockholm ligger inte @ Skåne.",
		`Strax utanför Amsterdam ligger en av Europas mest trafikerade flygplatser, Amsterdam-Schiphols flygplats, vars IATA-flygplatskod är "AMS"`,
		"Wisconsin är en delstat i USA, belägen i Mellanvästern och grundad som landets trettionde delstat 1848",
	}

	textSents := []text.Sentence{}
	for _, s := range sents {
		textSents = append(textSents, text.ComputeSentence(s))
	}

	a := text.Article{
		URL: "testgetbatch:testsource",
		Paragraphs: []text.Paragraph{
			{Sentences: textSents},
		},
	}

	// add sents
	aID, sentsWithID, err := Add(a, true)
	if err != nil {
		t.Errorf("Add went wrong : %v", err)
		return
	}
	if aID < 1 {
		t.Errorf("Got zero ID")
		return
	}

	insert := "INSERT INTO batch (chunk_id, name) VALUES (?,?)"
	n := 0
	for _, s := range sentsWithID {
		if contains(addToBatch, s.Text) {
			_, err = db.Exec(insert, s.ID, batchName)
			if err != nil {
				t.Errorf("add to batch failed : %v", err)
				return
			}
			n++
		}
	}
	if n != len(addToBatch) {
		t.Errorf("expected %d, got %d", len(addToBatch), n)
		return
	}

	// test pages
	var expect, got []string
	var script []text.Sentence

	var script2sents = func(script []text.Sentence) []string {
		res := []string{}
		for _, s := range script {
			res = append(res, s.Text)
		}
		return res
	}

	//
	expect = addToBatch
	script, err = GetBatch(batchName, 0, 0)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 6 {
		t.Errorf("Expected 6 sents in batch, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	//
	expect = addToBatch[0:5]
	script, err = GetBatch(batchName, 1, 5)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 5 {
		t.Errorf("Expected 5 sents in batch, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	//
	expect = []string{addToBatch[5]}
	script, err = GetBatch(batchName, 2, 5)
	if err != nil {
		t.Errorf("Couldn't get sents : %v", err)
		return
	}
	if len(script) != 1 {
		t.Errorf("Expected 1 sent in batch, found %d", len(script))
	}
	got = script2sents(script)
	if !reflect.DeepEqual(got, expect) {
		t.Errorf("Expected %v, got %v", expect, got)
	}
}
