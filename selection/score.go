package selection

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"

	"github.com/stts-se/manuscriptor2000/protocol"
	"github.com/stts-se/manuscriptor2000/text"
)

type Feat struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func AvailableFeats() []Feat {
	res := []Feat{
		{Name: text.FeatBigram,
			Desc: "Two-letter combinations",
		},
		{Name: text.FeatTrigram,
			Desc: "Three-letter combinations",
		},
		{Name: text.FeatBigram + "_top800",
			Desc: "The 800 most common bigrams",
		},
		{Name: text.FeatBigramTransition,
			Desc: "Bigrams in transititions between words",
		},
		{Name: text.FeatInitialBigram,
			Desc: "Sentence initial bigrams",
		},
		{Name: text.FeatFinalTrigram,
			Desc: "Sentence final trigrams",
		},
		{Name: text.FeatWord,
			Desc: "Words",
		},
	}
	//sort.Slice(res, func(i, j int) bool { return res[i] < res[j] })
	return res
}

var DefaultFeatureOpts = []protocol.SelectorFeatOpt{
	{text.FeatBigramTransition, 0},
	{text.FeatBigram + "_top800", 3},
	{text.FeatFinalTrigram, 0},
	{text.FeatInitialBigram, 0},
	{text.FeatWord, 0},
	{text.FeatBigram, 0},
}

// Chunk holds a chunk of sentences. A chunk is a set of sentences whose score is are computed as one entity, to speed up selection.
type Chunk struct {
	Sents           []Sent
	ScoreSet        ScoreSet
	SelectedFeature string
	DebugInfo       string
}

func (ch Chunk) SentStrings() []string {
	res := []string{}
	for _, s := range ch.Sents {
		res = append(res, fmt.Sprintf("%s\t%v\t%s\n", s.Sentence.Source, s.Sentence.ID, s.Sentence.Text))
	}
	return res
}

// MetaData returns a meta information string containing the chunk's score set, selected feature, and number of sentences
func (ch Chunk) MetaData() string {
	return fmt.Sprintf("%v\t%v\t%v", ch.ScoreSet, ch.SelectedFeature, len(ch.Sents))
}

// Text returns the text of each sentence in the chunk
func (ch Chunk) Text() []string {
	res := []string{}
	for _, s := range ch.Sents {
		res = append(res, s.Sentence.Text)
	}
	return res
}

func SumOfStats(options protocol.SelectorOptions, sents ...Sent) Stats {
	res := NewStats(options.FeatureOpts)
	for _, sent := range sents {
		res.Add(sent.Stats)
	}
	return res
}

type Sent struct {
	Sentence text.Sentence
	Stats    Stats
	// SelectedFeature string
	// ScoreSet        ScoreSet
	// DebugInfo       string
}

// Length returns the length (number of characters) in the sentence text
func (s Sent) Length() int {
	return len([]rune(s.Sentence.Text))
}

func (s Sent) String() string {
	//return fmt.Sprintf("%v\t%v\t%v\t%v", s.Sentence.ID, s.Sentence.Text, s.ScoreSet, s.SelectedFeature)
	return fmt.Sprintf("%v\t%v", s.Sentence.ID, s.Sentence.Text)
}

func StatsFromSent(s text.Sentence) Stats {
	res := Stats{} // NewStats()

	for fName, featMap := range s.Feats {
		res[fName] = FreqMap(featMap)
	}
	return res
}

// FreqMap holds frequencies for a selection feature
type FreqMap map[string]int

// Stats contains the numbers calculated for a sentence, such as
// character combination frequences
type Stats map[string]FreqMap

func NewStats(featOpts []protocol.SelectorFeatOpt) Stats {
	s := make(map[string]FreqMap)
	for _, opt := range featOpts {
		s[opt.Name] = make(map[string]int)
	}
	return Stats(s)
}

func (this Stats) Add(that Stats) {
	//requireStatsSymmetry(this)
	//requireStatsSymmetry(that)
	for name, thisM := range this {
		appendMap(thisM, that[name])
	}
}

func (this Stats) Copy() Stats {
	that := Stats{}
	for name, m := range this {
		that[name] = copyMap(m)
	}
	return that
}

// SetDiff returns a new Stats instance that only contains the stats of s2 that is not present in s1
func (s1 Stats) SetDiff(s2 Stats) Stats {
	//requireStatsSymmetry(s1)
	//requireStatsSymmetry(s2)
	res := Stats{}
	for name, m1 := range s1 {
		diff := setDiff(m1, s2[name])
		res[name] = diff
	}
	return res
}

// setDiff returns a new map instance that only contains the key value pairs in s2 that are not present in s1
func setDiff(m1, m2 map[string]int) map[string]int {
	res := map[string]int{}
	for k, v := range m2 {
		if _, ok := m1[k]; !ok {
			res[k] = v
		}
	}
	return res
}

// UpdateStats replaces the Stats struct of the sentence with the
// Stats entries that are not found in soFar
func UpdateStats(soFar Stats, ss []Sent) []Sent {
	var res []Sent
	for _, s := range ss {
		s.Stats = soFar.SetDiff(s.Stats)
		res = append(res, s)
	}
	return res
}

func (this Stats) FeatNames() []string {
	res := []string{}
	for k := range this {
		res = append(res, k)
	}
	return res
}

// Score holds a float64 value for a feature
type Score float64

// ScoreSet contains the scores calculated for a sentence (per feature score, e.g. bigram score, etc)
type ScoreSet map[string]Score

func NewScoreSet(featOpts []protocol.SelectorFeatOpt) ScoreSet {
	res := map[string]Score{}
	for name := range NewStats(featOpts) {
		res[name] = 0.0
	}
	return res
}

func (s ScoreSet) SortedKeys() []string {
	res := []string{}
	for k := range s {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (s ScoreSet) String() string {
	res := []string{}
	for _, k := range s.SortedKeys() {
		res = append(res, fmt.Sprintf("%s:%.3f", k, s[k]))
	}
	return strings.Join(res, " ")
}

func (s ScoreSet) IsZero() bool {
	var sum = Score(0.0)
	for _, score := range s {
		sum += score
	}
	return sum == 0.0
}

// IsHigherThan compare one score set to another, and return the name of the feature that determined the result
func (selector *Selector) IsHigherThan(this, that ScoreSet) (bool, string) {
	//requireScoreSetSymmetry(this)
	//requireScoreSetSymmetry(that)

	for _, opt := range selector.Options.FeatureOpts {
		s1 := roundScore(this[opt.Name])
		s2 := roundScore(that[opt.Name])
		if s1 != s2 {
			return s1 > s2, opt.Name
		}
	}
	return false, ""
}

func (selector *Selector) ScoreSet2String(this ScoreSet) string {
	//requireScoreSetSymmetry(this)
	res := []string{}
	for _, opt := range selector.Options.FeatureOpts {
		s := this[opt.Name]
		res = append(res, fmt.Sprintf("%v: %.4f", opt.Name, roundScore(s)))
	}
	return strings.Join(res, " ")
}

func (this ScoreSet) FeatNames() []string {
	res := []string{}
	for k := range this {
		res = append(res, k)
	}
	return res
}

//  round score to the nearest 0.0001 unit
const roundScoreFactor = 0.0001

func roundScore(s Score) Score {
	return Score(Round(float64(s), roundScoreFactor))
}

// getScore for set of sentence candidates, compared to the previously accumulated stats
func (selector *Selector) getScore(acc Stats, adjustScoreForSentenceLength bool, candidates ...Sent) ScoreSet {
	//requireStatsSymmetry(acc)
	//requireStatsSymmetry(candidate.Stats)

	var res = ScoreSet{}
	stats := SumOfStats(selector.Options, candidates...)
	for name, mAcc := range acc {
		diff := 0.0
		mCand := stats[name] // candidate.Stats[name]
		targetAmount := 0
		for _, opt := range selector.Options.FeatureOpts {
			if opt.Name == name {
				targetAmount = opt.TargetAmount
			}
		}
		for k := range mCand {
			if targetAmount > 0 {
				if mAcc[k] < targetAmount { // score = 1.0 if the previously accumulated frequency is under a certian frequency
					diff += 1.0
					continue
				} else {
					divBy := float64(mAcc[k] + 1) // nth finding: div by n (new feature: div by 1, second finding: div by 2, etc)
					d := (1.0 / divBy)
					diff += d
				}
			} else {
				divBy := float64(mAcc[k] + 1) // nth finding: div by n (new feature: div by 1, second finding: div by 2, etc)
				d := (1.0 / divBy)
				diff += d
			}
		}
		if adjustScoreForSentenceLength { // TODO MOVE UP TO SUM OF STATS?
			candidatesLen := 0
			for _, c := range candidates {
				candidatesLen += c.Length()
			}
			divBy := float64(candidatesLen) / 10
			diff = diff / divBy
		}
		res[name] = Score(diff)
		if selector.Options.Debug {
			log.Printf("getScore debug\t%s\t%v\t%v\t%f", name, mAcc, mCand, diff)
		}
	}
	return res
}

// MostNewInfo returns
// "best" next chunk
// the "best" next chunk's index
// a list of sentence indices that can be removed (since they contain no new information)
func (selector *Selector) MostNewInfo(selected Stats, sents []Sent, adjustScoreForSentenceLength bool) (Chunk, []int, map[int]bool) {
	removeIndices := make(map[int]bool)
	nSents := len(sents)
	if nSents == 0 {
		return Chunk{}, []int{}, removeIndices
	}

	type sentHolder struct {
		index int
		sent  Sent
		//score ScoreSet
	}
	type sentsHolder struct {
		sents []sentHolder
		score ScoreSet
	}
	var shSents = func(sh sentsHolder) []Sent {
		res := []Sent{}
		for _, s := range sh.sents {
			res = append(res, s.sent)
		}
		return res
	}
	type chanres struct {
		scores []sentsHolder
	}

	// var debugBestAndSecondBest = func(sents []sentsHolder) (ScoreSet, ScoreSet) {
	// 	res := []ScoreSet{}
	// 	for _, sh := range sents {
	// 		res = append(res, sh.score)
	// 	}
	// 	sort.Slice(res, func(i, j int) bool { r, _ := selector.IsHigherThan(res[i], res[j]); return r })
	// 	return res[0], res[1]
	// }

	var getScoreForAsyncSubset = func(accres chan chanres, ss []sentsHolder) {
		for i, sh := range ss {
			score := selector.getScore(selected, adjustScoreForSentenceLength, shSents(sh)...)
			sh.score = score
			ss[i] = sh
		}
		rchan := chanres{ss}
		accres <- rchan
	}
	var groupIntoChunks = func(shs []sentHolder) []sentsHolder {
		var res = []sentsHolder{}
		var acc = []sentHolder{}
		for i, sh := range shs {
			if i > 0 && i%selector.currentChunkSize == 0 && len(acc) > 0 {
				res = append(res, sentsHolder{sents: acc})
				acc = []sentHolder{}
			}
			acc = append(acc, sh)
		}
		if len(acc) > 0 {
			res = append(res, sentsHolder{sents: acc})
			acc = []sentHolder{}
		}
		return res
	}
	var computeScoresAsync = func() []sentsHolder {
		var accres = make(chan chanres)
		var nAsyncSubsets = 0
		var asyncSubsetSize = 2000
		var accInput = []sentHolder{}
		rand.Shuffle(len(sents), func(i, j int) { sents[i], sents[j] = sents[j], sents[i] })
		for i, s := range sents {
			if i > 0 && i%asyncSubsetSize == 0 && len(accInput) > 0 {
				nAsyncSubsets++
				chunks := groupIntoChunks(accInput)
				go getScoreForAsyncSubset(accres, chunks)
				accInput = []sentHolder{}
			}
			accInput = append(accInput, sentHolder{index: i, sent: s})
		}
		if len(accInput) > 0 {
			nAsyncSubsets++
			chunks := groupIntoChunks(accInput)
			go getScoreForAsyncSubset(accres, chunks)
			accInput = []sentHolder{}
		}

		res := []sentsHolder{}
		for i := 0; i < nAsyncSubsets; i++ {
			rr := <-accres
			for _, score := range rr.scores {
				res = append(res, score)
			}
		}
		//sort.Slice(res, func(i, j int) bool { return res[i].index < res[j].index })
		return res
	}

	scores := computeScoresAsync()

	best := Chunk{}
	bestIs := []int{}
	bestScore := NewScoreSet(selector.Options.FeatureOpts)
	for _, sh := range scores {
		score := sh.score
		if selector.Options.Debug {
			debugSents := Chunk{}
			for _, s := range sh.sents {
				debugSents.Sents = append(debugSents.Sents, s.sent)
			}
			log.Printf("MostNewInfo debug\t%s\t%s", debugSents.Text(), score)
		}
		if score.IsZero() {
			//removeIndices[sh.index] = true
			for _, s := range sh.sents {
				removeIndices[s.index] = true
			}
			continue
		}
		if isHigher, featName := selector.IsHigherThan(score, best.ScoreSet); isHigher {
			best.Sents = []Sent{}
			bestIs = []int{}
			for _, s := range sh.sents {
				best.Sents = append(best.Sents, s.sent)
				bestIs = append(bestIs, s.index)
			}
			best.SelectedFeature = featName
			best.ScoreSet = score
		}
	}
	if selector.Options.Debug {
		log.Printf("\nMostNewInfo debug BEST\t%s\t%s\t%s", best.Text(), bestScore, best.SelectedFeature)
	}
	return best, bestIs, removeIndices
}
