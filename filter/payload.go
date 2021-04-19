package filter

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/stts-se/manuscriptor2000/protocol"
)

const (
	WordCount      = "word_count"
	CommaCount     = "comma_count"
	SourceRE       = "source_re"
	ParagraphCount = "paragraph_count"
	SentenceCount  = "sentence_count"
	DigitCount     = "digit_count"
	LowestWordFreq = "lowest_word_freq"
	ExcludeChunkRE = "exclude_chunk_re"
	ChunkFeatCats  = "chunkfeat_cats"
	ExcludeBatches = "exclude_batches"
)

const (
	ChunkFeatCatsDocDesc    = "Choose sentences included in pre-defined feature categories (typically domains)"
	ChunkFeatCatsDocArgs    = "List of feature categories"
	ChunkFeatCatsDocExample = "se_place"
)

type Feat struct {
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Args    string `json:"args"`
	Example string `json:"example"`
}

func AvailableFeats() []Feat {
	res := []Feat{
		{
			Name:    CommaCount,
			Desc:    "Number of commas in a sentence",
			Args:    "Two integers defining a legal interval",
			Example: "0, 2",
		},
		{
			Name:    DigitCount,
			Desc:    "Number of digit expressions in a sentence",
			Args:    "Two integers defining a legal interval",
			Example: "0, 0",
		},
		{
			Name:    ExcludeBatches,
			Desc:    "Exclude sentences from certain batches",
			Args:    "List of batch names",
			Example: "test_batch_1",
		},
		{
			Name:    ExcludeChunkRE,
			Desc:    "Illegal sentence pattern",
			Args:    "Regular expression",
			Example: `[\p{Greek}]`,
		},
		{
			Name:    LowestWordFreq,
			Desc:    "Lowest word frequency allowed in a sentence",
			Args:    "One integer defining the frequency",
			Example: "3",
		},
		{
			Name:    ParagraphCount,
			Desc:    "Choose sentences from texts (sources) containing a certain number of paragraphs",
			Args:    "Two integers defining a legal interval",
			Example: "10, 20",
		},
		{
			Name:    SentenceCount,
			Desc:    "Choose sentences from texts (sources) containing a certain number of sentences",
			Args:    "Two integers defining a legal interval",
			Example: "9, -1",
		},
		{
			Name:    SourceRE,
			Desc:    "Required pattern for text source",
			Args:    "Regular expression",
			Example: "00$",
		},
		{
			Name:    WordCount,
			Desc:    "Number of words in a sentence",
			Args:    "Two integers defining a legal interval",
			Example: "4, 25",
		},
		{
			Name:    ChunkFeatCats,
			Desc:    ChunkFeatCatsDocDesc,
			Args:    ChunkFeatCatsDocArgs,
			Example: ChunkFeatCatsDocExample,
		},
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res
}

func args2int2(args []string) (int, int, error) {
	if len(args) != 2 {
		return 0, 0, fmt.Errorf("expected 2 args, found %d", len(args))
	}
	i1, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, 0, fmt.Errorf("couldn't parse int : %v", err)
	}
	i2, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, 0, fmt.Errorf("couldn't parse int : %v", err)
	}
	return i1, i2, err
}

func args2int(args []string) (int, error) {
	if len(args) != 1 {
		return 0, fmt.Errorf("expected 1 args, found %d", len(args))
	}
	i, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("couldn't parse int : %v", err)
	}
	return i, err
}

func args2string(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("expected 1 args, found %d", len(args))
	}
	return args[0], nil
}

func args2strings(args []string) ([]string, error) {
	return args, nil
}

func payloadOpt2filterOpt(o protocol.FilterOpt) (opt, error) {
	var res opt
	switch o.Name {
	case WordCount:
		i1, i2, err := args2int2(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		if i2 < 0 {
			return res, fmt.Errorf("cannot create filter with word count lower than %v", i2)
		}
		return wordCountView(i1, i2), nil
	case CommaCount:
		i1, i2, err := args2int2(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		if i2 < 0 {
			return res, fmt.Errorf("cannot create filter with comma count lower than %v", i2)
		}
		return commaCountView(i1, i2), nil
	case SourceRE:
		s, err := args2string(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return fromSourceRE(s), nil
	case ParagraphCount:
		i1, i2, err := args2int2(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		if i1 < 0 {
			return paragraphCountMax(i2), nil
		} else if i2 < 0 {
			return paragraphCountMin(i1), nil
		} else {
			return paragraphCountInterval(i1, i2), nil
		}
	case SentenceCount:
		i1, i2, err := args2int2(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		if i1 < 0 {
			return sentenceCountMax(i2), nil
		} else if i2 < 0 {
			return sentenceCountMin(i1), nil
		} else {
			return sentenceCountInterval(i1, i2), nil
		}
	case DigitCount:
		i, err := args2int(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return nDigitCountView(i), nil
	case LowestWordFreq:
		i, err := args2int(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return lowestWordFreqCountView(i), nil
	case ExcludeChunkRE:
		s, err := args2string(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return excludeChunkRE(s), nil
	case ExcludeBatches:
		ss, err := args2strings(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return tailNotInBatches(ss...), nil
	case ChunkFeatCats:
		ss, err := args2strings(o.Args)
		if err != nil {
			return res, fmt.Errorf("couldn't parse %s opt : %v", o.Name, err)
		}
		return chunkFeatCat(ss...), nil
	default:
		return res, fmt.Errorf("unknown option type: %s", o.Name)
	}
}

func NewQueryBuilder(payload protocol.FilterPayload) (*queryBuilder, error) {
	filterOpts := []opt{
		filterQHeadInto(payload.BatchName),
	}

	seenExcludeBatches := false
	for _, o0 := range payload.Opts {
		o, err := payloadOpt2filterOpt(o0)
		if err != nil {
			return &queryBuilder{}, fmt.Errorf("couldn't build filter opt from %v : %v", o0, err)
		}
		filterOpts = append(filterOpts, o)
		if o0.Name == ExcludeBatches {
			seenExcludeBatches = true
		}
	}
	if !seenExcludeBatches {
		filterOpts = append(filterOpts, tailNotInBatches())
	}

	if payload.TargetSize > 0 {
		filterOpts = append(filterOpts, tailLimit(payload.TargetSize))
	}
	return newFilterQueryBuilder(filterOpts...)

}
