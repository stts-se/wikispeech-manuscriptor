package selection

import (
	"reflect"
	"testing"

	"github.com/stts-se/manuscriptor2000/protocol"
	"github.com/stts-se/manuscriptor2000/text"
)

func TestSumOfStats(t *testing.T) {
	opts := protocol.SelectorOptions{
		FeatureOpts: DefaultFeatureOpts,
	}
	texts := []string{
		"Jag är en apa",
		"Vintern kommer snart, erkände Thomas.",
		"Soporna behöver tömmas.",
		"Solen skiner.",
		"sommaren slutar, nu!",
	}
	sents := []Sent{}
	for _, txt := range texts {
		s0 := text.ComputeSentence(txt)
		sent := Sent{Sentence: s0, Stats: StatsFromSent(s0)}
		sents = append(sents, sent)
		//fmt.Printf("%#v\n", sent.Stats)
	}

	res := SumOfStats(opts, sents...)

	expect := Stats{
		text.FeatFinalTrigram: FreqMap{
			"apa": 1,
			"mas": 2,
			"ner": 1,
			"rnu": 1,
		},
		text.FeatInitialBigram: FreqMap{
			"ja": 1,
			"vi": 1,
			"so": 3,
		},
	}

	testFeatNames := []string{text.FeatFinalTrigram, text.FeatInitialBigram}
	for _, fn := range testFeatNames {
		if !reflect.DeepEqual(res[fn], expect[fn]) {
			t.Errorf("Expected %v, found %v", expect[fn], res[fn])
		}
	}
}
