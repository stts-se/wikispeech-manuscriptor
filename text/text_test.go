package text

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStiTokens(t *testing.T) {

	s1 := "?ø"

	toks := s2Tokens(s1)

	if w, g := 2, len(toks); w != g {
		t.Errorf("wanted '%d' got '%d'", w, g)
	}

	if w, g := FeatPunct, toks[0].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "?", toks[0].Text; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	s2 := "Amager"

	toks = s2Tokens(s2)

	if w, g := 1, len(toks); w != g {
		t.Errorf("wanted '%d' got '%d'", w, g)
	}

	if w, g := "letter", toks[0].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "Amager", toks[0].Text; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	s3 := "aaa  (bbb) ccc"

	toks = s2Tokens(s3)

	if w, g := 7, len(toks); w != g {
		t.Errorf("wanted '%d' got '%d'", w, g)
	}

	if w, g := "letter", toks[0].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "aaa", toks[0].Text; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "space", toks[1].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "  ", toks[1].Text; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := FeatPunct, toks[2].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "letter", toks[6].Type; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	if w, g := "ccc", toks[6].Text; w != g {
		t.Errorf("wanted '%s' got '%s'", w, g)
	}

	s4 := `a?!"`

	toks = s2Tokens(s4)
	if w, g := 4, len(toks); w != g {
		t.Errorf("wanted '%d' got '%d'", w, g)
	}

}

func TestSentence(t *testing.T) {

	s1 := ComputeSentence("Four tokens?")
	if w, g := 1, len(s1.Feats[FeatPunct]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 2, len(s1.Feats[FeatWord]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 1, s1.Feats[FeatWord]["four"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	s2 := ComputeSentence("Six six tokens!")

	// if w, g := 6, len(s2.Tokens); w != g {
	// 	t.Errorf("wanted %d got %d", w, g)
	// }

	if w, g := 2, len(s2.Feats[FeatWord]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 2, s2.Feats[FeatWord]["six"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	s3 := ComputeSentence("Seven seven tokens!!")

	if w, g := 1, len(s3.Feats[FeatPunct]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 2, s3.Feats[FeatPunct]["!"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 2, len(s3.Feats[FeatWord]); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 2, s3.Feats[FeatWord]["seven"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

}

func blee() { fmt.Println() }

func TestAdd(t *testing.T) {

	m1 := map[string]int{"w1": 57, "w2": 1, "w3": 3}
	m2 := map[string]int{"w1": 1, "w3": 3, "w4": 11}
	m3 := map[string]int{"w1": 3, "w4": 1, "w5": 79}

	res := Add(m1, m2, m3)

	if w, g := 61, res["w1"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 1, res["w2"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 6, res["w3"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 12, res["w4"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := 79, res["w5"]; w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

}

func TestBigramTransitionsStrn(t *testing.T) {

	b1 := BigramTransitionsStrn("hej du")
	if w, g := 1, len(b1); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	b2 := BigramTransitionsStrn("hej du strut")
	if w, g := 2, len(b2); w != g {
		t.Errorf("wanted %d got %d", w, g)
	}

	if w, g := "j d", b2[0]; w != g {
		t.Errorf("wanted %s got %s", w, g)
	}

	if w, g := "u s", b2[1]; w != g {
		t.Errorf("wanted %s got %s", w, g)
	}

}

func TestNGrams(t *testing.T) {

	s := "'du är en gammal häst', sa Nisse 14 gånger, ohörbart."
	tokens := s2Tokens(s)

	monos := NGrams(tokens, 1)
	expMonos := []string{"d", "u", "ä", "r", "e", "n", "g", "a", "m", "m", "a", "l", "h", "ä", "s", "t", "s", "a", "n", "i", "s", "s", "e", "g", "å", "n", "g", "e", "r", "o", "h", "ö", "r", "b", "a", "r", "t"}
	if !reflect.DeepEqual(monos, expMonos) {
		t.Errorf("Expected\n%#v, got\n%#v", expMonos, monos)
	}

	bis := NGrams(tokens, 2)
	expBis := []string{"du", "uä", "är", "re", "en", "ng", "ga", "am", "mm", "ma", "al", "lh", "hä", "äs", "st", "sa", "an", "ni", "is", "ss", "se", "gå", "ån", "ng", "ge", "er", "oh", "hö", "ör", "rb", "ba", "ar", "rt"}
	if !reflect.DeepEqual(bis, expBis) {
		t.Errorf("Expected\n%#v, got\n%#v", expBis, bis)
	}

}

func TestInitialNGram(t *testing.T) {

	var exp, got string

	s := "'du är en gammal häst', sa Nisse 14 gånger, ohörbart."
	tokens := s2Tokens(s)

	exp = "d"
	got = InitialNGram(tokens, 1)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

	exp = "du"
	got = InitialNGram(tokens, 2)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

	exp = "duäre"
	got = InitialNGram(tokens, 5)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

}

func TestFinalNGram(t *testing.T) {

	var exp, got string

	s := "'du är en gammal häst', sa Nisse 14 gånger, ohörbart jo."
	tokens := s2Tokens(s)

	exp = "o"
	got = FinalNGram(tokens, 1)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

	exp = "jo"
	got = FinalNGram(tokens, 2)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

	exp = "tjo"
	got = FinalNGram(tokens, 3)
	if got != exp {
		t.Errorf("Expected\n%#v, got\n%#v", exp, got)
	}

}
