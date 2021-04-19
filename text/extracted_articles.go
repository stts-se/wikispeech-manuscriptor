package text

import (
	"regexp"
	"strings"
)

func string2Article(s string) Article {
	var res Article
	s = strings.TrimSpace(s)
	paras := strings.Split(s, "\n\n")
	//paras := strings.Split(s, "\n")

	// First line: doc tag
	// Second line: title word
	if len(paras) < 2 {
		return res
	}

	// First paragraph expected to be opening doc tag with url, plus title
	docTag := paras[0]
	docSplt := strings.Split(docTag, " ")
	for _, ds := range docSplt {
		ds = strings.TrimSpace(ds)
		if strings.HasPrefix(ds, `url="`) {
			ds = strings.TrimPrefix(ds, `url="`)
			ds = strings.TrimSuffix(ds, `"`)
			res.URL = ds
		}
		if strings.HasPrefix(ds, `title="`) {
			ds = strings.TrimPrefix(ds, `title="`)
			ds = strings.TrimSuffix(ds, `"`)
			ds = strings.TrimSuffix(ds, `">`)
			res.Title = ds
		}

	}

	// First line: doc tag
	// Second line: title word
	for _, p := range paras[1:] {

		sents := string2Sentences(p)
		if len(sents) > 0 {
			newP := Paragraph{Sentences: sents}
			res.Paragraphs = append(res.Paragraphs, newP)
		}
	}

	return res
}

// TODO: DUMMY SENTENCE SPLIT
// \p{Ll}: Lowercase letter
// \p{P}: Punctuation
// \p{Lu}: Uppercase letter

// OLD
// var sentSplitRE = regexp.MustCompile(`\p{Ll}[”"]?[.!?]+[”"]?(\s+)["]?\p{Lu}`)

// NEW
var sentSplitRE = regexp.MustCompile(`[\p{Ll}0-9][”")]?[.!?]+[”")]?(\s+)["]?\p{Lu}`)

func string2Sentences(s string) []Sentence {
	// remove newlines and multiple spaces
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\n", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)

	var res []Sentence

	matchIndxs := sentSplitRE.FindAllStringSubmatchIndex(s, -1)

	start := 0

	for _, is := range matchIndxs {

		sent := s[start:is[2]]

		start = is[3]

		res = append(res, ComputeSentence(sent))

	}

	if start < len(s)-1 {
		lastSent := s[start:]
		res = append(res, ComputeSentence(lastSent))
	}

	return res
}

// ExtractedFile2Articles processes the outpuf of WikiExtrator.py
func ExtractedFile2Articles(s string) []Article {
	var res []Article

	articles := strings.Split(s, "</doc>")
	for _, s := range articles {

		a := string2Article(s)

		if a.URL != "" {

			res = append(res, a)
		}
	}

	return res
}
