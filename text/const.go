package text

const (
	FeatBigram           = "bigram"
	FeatTrigram          = "trigram"
	FeatBigramTransition = "bigram_transition"
	FeatFinalTrigram     = "final_trigram"
	FeatInitialBigram    = "initial_bigram"

	FeatWord              = "word"
	FeatCount             = "count"
	FeatValWordCount      = "word_count"
	FeatValSentenceCount  = "sentence_count"
	FeatValParagraphCount = "paragraph_count"
	FeatValDigitCount     = "digit_count"
	FeatValLowestWordFreq = "lowest_word_freq"
	FeatPunct             = "punct"
	FeatSEPlace           = "se_place"
	FeatSESurname         = "se_surname"
	FeatSEFemName         = "se_fem_name"
	FeatSEMaleName        = "se_male_name"

	// BlockBatch is a special batch used for blocking sentences from filtering/selection. All filter/selection queries will exclude this batch.
	BlockBatch = "blocked"
)
