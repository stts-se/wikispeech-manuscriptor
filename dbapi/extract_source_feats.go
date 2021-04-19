package dbapi

import (
	//"strings"

	"github.com/stts-se/manuscriptor2000/text"
)

type SourceExtractor interface {
	Extract(text.Article) []Feat
}

type SourceFeaturesExtractor struct {
}

func (s SourceFeaturesExtractor) Extract(a text.Article) []Feat {

	var res []Feat

	pN := Feat{Name: text.FeatCount, Value: text.FeatValParagraphCount, Freq: len(a.Paragraphs)}

	n := 0
	for _, p := range a.Paragraphs {
		n += len(p.Sentences)
	}
	sN := Feat{Name: text.FeatCount, Value: text.FeatValSentenceCount, Freq: n}

	res = append(res, pN, sN)

	return res
}
