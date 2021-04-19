package selection

import (
	"fmt"
	//"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
	"github.com/stts-se/wikispeech-manuscriptor/protocol"
	"github.com/stts-se/wikispeech-manuscriptor/text"
)

const (
	ModeRand       = "rand"
	ModeExhaustive = "exhaustive"
)

func Validate(o protocol.SelectorOptions) error {
	if o.TargetSize < 1 {
		return fmt.Errorf("target amount must be more than zero, found %v", o.TargetSize)
	}
	if o.ScriptName == "" {
		return fmt.Errorf("script name not provided")
	}
	if o.FromBatch == "" {
		return fmt.Errorf("input batch not provided")
	}
	return nil
}

type Selector struct {
	Options                 protocol.SelectorOptions
	Corpus                  []Sent
	InputBatchStats         Stats
	InputBatchSize          int
	AccumulatedScriptsStats Stats
	AccumulatedScriptsSize  int
	Selection               []Sent
	SelectionStats          Stats
	currentChunkSize        int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewSelector creates a new Selector instance based on the input options
func NewSelector(options protocol.SelectorOptions) (Selector, error) {
	switch options.Mode {
	case ModeRand:
	case ModeExhaustive:
	default:
		return Selector{}, fmt.Errorf("unknown selector mode: %s", options.Mode)
	}

	var res = Selector{
		Options:                options,
		Corpus:                 []Sent{},
		Selection:              []Sent{},
		InputBatchSize:         0,
		AccumulatedScriptsSize: 0,
	}
	res.SelectionStats = NewStats(options.FeatureOpts)
	res.InputBatchStats = NewStats(options.FeatureOpts)
	res.AccumulatedScriptsStats = NewStats(options.FeatureOpts)
	res.currentChunkSize = options.ChunkSize

	return res, nil
}

func (selector *Selector) TargetReached() bool {
	if len(selector.Selection) >= selector.Options.TargetSize {
		return true
	}
	return false
}

var querySelectFromBatch = `SELECT chunk.id FROM chunk, batch WHERE batch.name = ? AND batch.chunk_id = chunk.id AND chunk.id NOT IN ( SELECT chunk_id from script WHERE script.name IN (%s) ) AND chunk.id NOT IN ( SELECT chunk_id from batch where batch.name = '` + text.BlockBatch + `')`

const querySelectFromAccScripts = `SELECT chunk.id FROM chunk, script WHERE script.chunk_id = chunk.id AND chunk.id IN ( SELECT chunk_id from script WHERE script.name IN (%s) ) AND chunk.id NOT IN ( SELECT chunk_id from batch where batch.name = '` + text.BlockBatch + `')`

func (selector *Selector) getInputBatch() ([]Sent, error) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[selector] Computing input batch data from batch %s ... ", selector.Options.FromBatch)
	selector.InputBatchStats = NewStats(selector.Options.FeatureOpts)
	var res []Sent

	// build query
	qs := []string{}
	args := []interface{}{selector.Options.FromBatch}
	for _, scriptName := range selector.Options.AccumulatedScripts {
		qs = append(qs, "?")
		args = append(args, scriptName)
	}
	query := fmt.Sprintf(querySelectFromBatch, strings.Join(qs, ","))
	//log.Printf("selector.getInputBatch debug populatedQuery\t%s", populateQueryString(query, args))

	// exec query
	tx, err := dbapi.Begin()
	if err != nil {
		return res, fmt.Errorf("couldn't create db connection : %v", err)
	}
	rows, err := tx.Query(query, args...)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("getInputBatch failed to query db : %v", err)
	}

	ids := []int64{}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			tx.Rollback()
			return res, fmt.Errorf("failed to scan rows : %v", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading result row : %v", err)
	}
	rows.Close() // strictly not needed
	err = tx.Commit()
	if err != nil {
		return res, fmt.Errorf("couldn't commit transaction : %v", err)
	}

	//fmt.Fprintf(os.Stderr, "\n - ids fetched\n")

	// shuffle needed for chunked selection
	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	//sents, err := dbapi.GetSents(ids...)
	sents, err := dbapi.GetSentsInBatches(20000, ids...)
	if err != nil {
		return res, fmt.Errorf("failed to get sents : %v", err)
	}
	//fmt.Fprintf(os.Stderr, "\n - sents fetched\n")
	for _, s := range sents {
		// if i%100 == 0 {
		// 	fmt.Fprintf(os.Stderr, "!")
		// }
		sent := Sent{Sentence: s, Stats: StatsFromSent(s)}
		selector.InputBatchStats.Add(sent.Stats)
		res = append(res, sent)
	}

	selector.InputBatchSize = len(res)
	dur := time.Since(start)
	fmt.Fprintf(os.Stderr, "done\n")
	fmt.Fprintf(os.Stderr, "[selector] Loaded %d sentences from batch %s (took %v)\n", len(res), selector.Options.FromBatch, dur)
	return res, nil
}

func (selector *Selector) loadAccumulatedScripts() error {
	if len(selector.Options.AccumulatedScripts) == 0 {
		fmt.Fprintf(os.Stderr, "[selector] No accumulated script data to compute\n")
		return nil
	}
	fmt.Fprintf(os.Stderr, "[selector] Computing accumulated scripts data from %s ... ", strings.Join(selector.Options.AccumulatedScripts, ","))
	selector.AccumulatedScriptsStats = NewStats(selector.Options.FeatureOpts)

	// build query
	qs := []string{}
	args := []interface{}{}
	for _, scriptName := range selector.Options.AccumulatedScripts {
		qs = append(qs, "?")
		args = append(args, scriptName)
	}
	query := fmt.Sprintf(querySelectFromAccScripts, strings.Join(qs, ","))
	//log.Printf("selector.loadAccumulatedScripts debug populatedQuery\t%s", populateQueryString(query, args))

	// exec query
	tx, err := dbapi.Begin()
	if err != nil {
		return fmt.Errorf("couldn't create db connection : %v", err)
	}
	rows, err := tx.Query(query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("getInputBatch failed to query db : %v", err)
	}

	ids := []int64{}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to scan rows : %v", err)
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("error when reading result row : %v", err)
	}
	rows.Close() // strictly not needed
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit transaction : %v", err)
	}

	sents, err := dbapi.GetSents(ids...)
	if err != nil {
		return fmt.Errorf("failed to get sents : %v", err)
	}
	for _, s := range sents {
		sent := Sent{Sentence: s, Stats: StatsFromSent(s)}
		selector.AccumulatedScriptsStats.Add(sent.Stats)
		selector.AccumulatedScriptsSize++
	}
	fmt.Fprintf(os.Stderr, "done\n")
	fmt.Fprintf(os.Stderr, "[selector] Loaded %d sentences from accumulated scripts %s\n", selector.AccumulatedScriptsSize, selector.Options.AccumulatedScripts)
	return nil
}

func (selector *Selector) Init() error {
	batch, err := selector.getInputBatch()
	selector.Corpus = batch
	if err != nil {
		return fmt.Errorf("couldn't get retreive input batch data : %v", err)
	}
	err = selector.loadAccumulatedScripts()
	if err != nil {
		return fmt.Errorf("couldn't get retrieve accumulated scripts data : %v", err)
	}
	return nil
}

func (selector *Selector) SelectNext() (string, bool) {
	switch selector.Options.Mode {
	case ModeRand:
		return selector.SelectNextRand()
	case ModeExhaustive:
		return selector.SelectNextExhaustive()
	default:
		panic(fmt.Sprintf("unknown selector mode: %s", selector.Options.Mode))
	}
}

func (selector *Selector) SelectNextRand() (string, bool) {
	var minIterations = selector.Options.MinIterationsRand
	var cutoff = selector.Options.CutoffRand

	var bestSoFar = NewScoreSet(selector.Options.FeatureOpts)
	var bestFeature string

	var iterations int
	var iterationsWithoutImprovement int

	var winner []Sent
	var winnerIndices []int

	logInterval := (minIterations / 20)
	for true {
		if iterations > 0 && iterations%logInterval == 0 || iterations == 1 {
			//log.Printf("SelectNextRand iteration %d, iterations without improvements %d", iterations, iterationsWithoutImprovement)
			//log.Printf("best score so far %v", bestSoFar)
			fmt.Fprintf(os.Stderr, "\rIterations %d|%d\r", iterations, iterationsWithoutImprovement)
		}
		iterations++
		if len(selector.Corpus) < selector.Options.TargetSize {
			return fmt.Sprintf("cannot select %d sententes from a corpus of %d", selector.Options.TargetSize, len(selector.Corpus)), false
		}
		rand.Shuffle(len(selector.Corpus), func(i, j int) { selector.Corpus[i], selector.Corpus[j] = selector.Corpus[j], selector.Corpus[i] })
		sents := selector.Corpus[0:selector.Options.TargetSize]

		score := selector.getScore(selector.SelectionStats, selector.Options.AdjustScoreForSentenceLength, sents...)
		foundBetter, feature := selector.IsHigherThan(score, bestSoFar)
		if foundBetter {
			bestSoFar = score
			bestFeature = feature
			winner = sents
			winnerIndices = []int{}
			for i := range sents {
				winnerIndices = append(winnerIndices, i)
			}
			iterationsWithoutImprovement = 0
		} else {
			iterationsWithoutImprovement++
		}
		if iterations >= minIterations && iterationsWithoutImprovement >= cutoff {
			break
		}
	}

	chunk := Chunk{ScoreSet: bestSoFar, Sents: winner, SelectedFeature: bestFeature}
	selector.cacheSelection(chunk, winnerIndices, make(map[int]bool))
	return fmt.Sprintf("%d iterations", iterations), true
}

func (selector *Selector) SelectNextExhaustive() (string, bool) {
	if len(selector.Corpus) > 0 {
		selected, selectedIndices, removeIndices := selector.MostNewInfo(selector.SelectionStats, selector.Corpus /*selector.requiredValues,*/, selector.Options.AdjustScoreForSentenceLength)

		//if selectedIndices < 0 {
		if len(selectedIndices) == 0 {
			return "unable to make a selection", false
		}
		if selected.ScoreSet.IsZero() {
			return "no new info in corpus", false
		}

		selector.cacheSelection(selected, selectedIndices, removeIndices)
		if selector.currentChunkSize > 1 {
			selector.currentChunkSize -= selector.Options.ChunkDecrease
		}
		return "", true
	}
	return "end of corpus", false
}

// returns number of deleted sentences
func (selector *Selector) removeCorpusIndices(indices map[int]bool) int {
	corpusUpdate := []Sent{}
	nDeleted := 0

	for i, sent := range selector.Corpus {
		if _, ok := indices[i]; !ok {
			corpusUpdate = append(corpusUpdate, sent)
		} else {
			nDeleted++
			// if selector.Options.Verb {
			// 	debug(fmt.Sprintf("REMOVED\t%v", sent))
			// }
		}
	}
	selector.Corpus = corpusUpdate
	return nDeleted
}

func (selector *Selector) cacheSelection(sents Chunk, sentIndexInCorpus []int, removableIndices map[int]bool) error {
	// if selector.Options.Verb && len(removableIndices) > 0 {
	// 	debug(fmt.Sprintf("REMOVING INDICES\t%v", removableIndices))
	// }

	for _, sent := range sents.Sents {
		selector.Selection = append(selector.Selection, sent)
	}

	lenBefore := len(selector.Corpus)

	// remove selected sentence and removable sentences from the corpus
	//if sentIndexInCorpus >= 0 {
	for _, i := range sentIndexInCorpus {
		//removableIndices[sentIndexInCorpus] = true
		removableIndices[i] = true
	}
	selector.removeCorpusIndices(removableIndices)
	lenAfter := len(selector.Corpus)
	if len(removableIndices) > 0 && lenAfter == lenBefore {
		return fmt.Errorf("corpus should have been reduced by %d sentences; still %d sentences left; sent index was %d", len(removableIndices), lenAfter, sentIndexInCorpus)
	}

	selector.SelectionStats.Add(SumOfStats(selector.Options, sents.Sents...))
	// if selector.Options.ContinuousPrint {
	// 	fmt.Println()
	// 	if selector.Options.PrintMetaData {
	// 		fmt.Printf("metadata: %v\n", sents.MetaData())
	// 	}
	// 	fmt.Println(strings.Join(sents.SentStrings(), "\n"))
	// }
	return nil
}

func (selector *Selector) WriteScriptToDB(metadata protocol.ScriptMetadata) (int, error) {
	if metadata.Timestamp == "" {
		metadata.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	ids := []int64{}
	for _, s := range selector.Selection {
		ids = append(ids, s.Sentence.ID)
	}

	return dbapi.SaveScript(metadata, ids...)
}
