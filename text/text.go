package text

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Token struct {
	//Name string
	Type string
	Text string
}

type Sentence struct {
	ID     int64  `json:"id"`
	Text   string `json:"text"`
	Source string `json:"source"`
	//Words map[string]int // TODO Nuke this and use only Feats
	// TODO: Feats should be map[featName]map[FeatVals]freq
	Feats map[string]map[string]int `json:"feats,omitempty"` // Todo Value should be FeatVal{Name string, Freq int)?
	//Tokens []Token
}

func (s *Sentence) AddFeat(name, value string) {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	_, ok := s.Feats[name]
	if !ok {
		s.Feats[name] = map[string]int{}
	}
	s.Feats[name][value]++
}

func (s *Sentence) AddFeatWithFreq(name, value string, freq int) {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	_, ok := s.Feats[name]
	if !ok {
		s.Feats[name] = map[string]int{}
	}
	s.Feats[name][value] = freq
}

type Paragraph struct {
	Sentences []Sentence
}

type Article struct {
	URL        string
	Title      string
	Paragraphs []Paragraph
}

func runeType(r rune) string {
	switch {
	case unicode.IsLetter(r):
		return "letter"
	case unicode.IsDigit(r):
		return "digit"
	case unicode.IsPunct(r):
		return "punct"
	case unicode.IsSpace(r):
		return "space"

	case unicode.IsSymbol(r):
		return "symbol"

	default:
		return "other"
	}
}

// A token is a sequence of runes of same rune type
func s2Tokens(s string) []Token {
	var res []Token
	var lastRuneType string

	lastToken := Token{}

	for _, r := range s {

		rType := runeType(r)

		// first rune
		if lastRuneType == "" {
			lastToken.Type = rType
			lastToken.Text = lastToken.Text + string(r)

			// same type of rune as last one, but each punct is its own token
		} else if lastRuneType == rType && rType != "punct" {

			lastToken.Text = lastToken.Text + string(r)
		} else {
			// new type of rune
			res = append(res, lastToken)
			lastToken = Token{
				Type: rType,
				Text: string(r),
			}
		}
		lastRuneType = rType
	}

	res = append(res, lastToken)

	return res
}

// ComputeSentence creates a new sentence and computes its feats
func ComputeSentence(s string) Sentence {
	res := Sentence{Text: s}
	res.Feats = make(map[string]map[string]int)
	tokens := s2Tokens(s)

	for _, t := range tokens {
		if t.Type == "space" {
			continue
		}
		if t.Type == "letter" {
			//res.Words[strings.ToLower(t.Text)]++
			res.AddFeat(FeatWord, strings.ToLower(t.Text))
		} else {
			res.AddFeat(t.Type, strings.ToLower(t.Text))
		}
	}

	for _, chBigramTr := range BigramTransitions(tokens) {
		res.AddFeat(FeatBigramTransition, chBigramTr)
	}
	for _, chBigram := range NGrams(tokens, 2) {
		res.AddFeat(FeatBigram, chBigram)
	}
	if bg := InitialNGram(tokens, 2); bg != "" {
		res.AddFeat(FeatInitialBigram, bg)
	}
	if tg := FinalNGram(tokens, 3); tg != "" {
		res.AddFeat(FeatFinalTrigram, tg)
	}
	// for _, chTrigram := range NGrams(tokens, 3) {
	// 	res.AddFeat(FeatTrigram, chTrigram)
	// }

	var nWords int
	for _, freq := range res.Feats[FeatWord] {
		nWords += freq
	}
	res.AddFeatWithFreq(FeatCount, FeatValWordCount, nWords)

	var nDigits int
	for _, freq := range res.Feats["digit"] {
		nDigits += freq
	}
	res.AddFeatWithFreq(FeatCount, FeatValDigitCount, nDigits)

	return res
}

func Add(ms ...map[string]int) map[string]int {
	res := make(map[string]int)

	for _, m := range ms {
		for k, v := range m {
			res[k] = res[k] + v
		}

	}

	return res
}

func NLastChars(s string, n int) string {
	r := []rune(s)
	return string(r[len(r)-n:])
}

func NFirstChars(s string, n int) string {
	r := []rune(s)

	return string(r[0:n])
}

func BigramTransitionsStrn(sent string) []string {
	return BigramTransitions(s2Tokens(sent))
}

func BigramTransitions(sent []Token) []string {
	var res []string
	var prev Token

	for _, t := range sent {

		if t.Type == "space" {
			continue
		}

		// "letter" means 'word' (letter sequence...)
		if prev.Type == "letter" && t.Type == "letter" {
			bi1 := NLastChars(prev.Text, 1)
			bi2 := NFirstChars(t.Text, 1)
			res = append(res, strings.ToLower(bi1+" "+bi2))
		}

		prev = t
	}

	return res
}

func getNGrams(s string, window int) []string {
	runes := []rune(strings.ToLower(s))
	res := []string{}
	for n := window; n <= len(runes); n++ {
		ngram := string(runes[n-(window) : n])
		res = append(res, ngram)
	}
	return res
}

func NGramsStrn(sent string, n int) []string {
	return NGrams(s2Tokens(sent), n)
}

func NGrams(sent []Token, n int) []string {
	var res []string
	var acc []string

	for _, t := range sent {

		if t.Type == "space" {
			continue
		}
		// "letter" means 'word' (letter sequence...)
		if t.Type == "letter" {
			acc = append(acc, strings.ToLower(t.Text))
		} else if t.Type != "letter" && len(acc) > 0 {
			ngrams := getNGrams(strings.Join(acc, ""), n)
			for _, ng := range ngrams {
				res = append(res, ng)
			}
			acc = []string{}
		}
	}

	return res
}

func InitialNGram(sent []Token, n int) string {
	if len(sent) == 0 {
		return ""
	}
	acc := []string{}
	for _, t := range sent {
		if t.Type == "letter" {
			acc = append(acc, strings.ToLower(t.Text))
		}
	}
	text := strings.Join(acc, "")
	runes := []rune(text)
	if len(runes) >= n {
		return string(runes[0:n])
	}
	return ""
}

func InitialNGramStrn(sent string, n int) string {
	return InitialNGram(s2Tokens(sent), n)
}

func FinalNGramStrn(sent string, n int) string {
	return FinalNGram(s2Tokens(sent), n)
}

func FinalNGram(sent []Token, n int) string {
	if len(sent) == 0 {
		return ""
	}
	acc := []string{}
	for _, t := range sent {
		if t.Type == "letter" {
			acc = append(acc, strings.ToLower(t.Text))
		}
	}
	text := strings.Join(acc, "")
	runes := []rune(text)
	l := len(runes)
	if len(runes) >= n {
		return string(runes[l-n : l])
	}
	return ""
}

const randomCharSetBase = "ABCDEFGHIJKLMNOPQRSTUVZYZabcdefghijklmnopqrstuvzyz0123456789_"

func RandomString(length int) string {
	var output strings.Builder
	runes := []rune(randomCharSetBase)
	for i := 0; i < length; i++ {
		random := rand.Intn(len(runes))
		randomChar := runes[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}
