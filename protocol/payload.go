package protocol

import (
	"encoding/json"
)

// type Message struct {
// 	Level string `json:"level,omitempty"`
// 	Text  string `json:"text,omitempty"`
// }

// type Payload struct {
// 	Name    string  `json:"name,omitempty"`
// 	Message Message `json:"message,omitempty"`
// }

type FilterPayload struct {
	//Payload
	BatchName  string      `json:"batch_name"`
	TargetSize int         `json:"target_size"`
	Opts       []FilterOpt `json:"opts"`
}

func (f FilterPayload) Empty() bool {
	return f.BatchName == "" && f.TargetSize == 0 && len(f.Opts) == 0
}

type BatchMetadata struct {
	FilterPayload
	OutputSize int    `json:"output_size"`
	Timestamp  string `json:"timestamp"`
}

func (meta BatchMetadata) Empty() bool {
	return meta.Timestamp == "" && meta.OutputSize == 0 && meta.BatchName == "" && len(meta.Opts) == 0
}

type FilterOpt struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// Selector

type SelectorFeatOpt struct {
	Name         string `json:"name"`
	TargetAmount int    `json:"target_amount"`
}

type SelectorOptions struct {

	// Mode: Selection mode (exhaustive or rand)
	Mode                         string            `json:"mode"`
	FeatureOpts                  []SelectorFeatOpt `json:"feature_opts"`
	AdjustScoreForSentenceLength bool              `json:"adjust_score_for_sentence_length"`
	TargetSize                   int               `json:"target_size"`
	FromBatch                    string            `json:"from_batch"`
	ScriptName                   string            `json:"script_name"`
	AccumulatedScripts           []string          `json:"accumulated_scripts"`
	//ContinuousPrint              bool              `json:"continuous_print,omitempty"`
	PrintMetaData bool `json:"print_metadata,omitempty"`

	// exhaustive search
	ChunkSize     int `json:"chunk_size"`
	ChunkDecrease int `json:"chunk_decrease"`

	// random search
	MinIterationsRand int `json:"min_iterations"`
	CutoffRand        int `json:"cutoff"`

	Debug bool `json:"debug,omitempty"`
}

func (s SelectorOptions) Empty() bool {
	return s.ScriptName == "" && s.TargetSize == 0 && len(s.FeatureOpts) == 0
}

type SelectorPayload struct {
	//Payload
	Options SelectorOptions `json:"options"`
}

type Config struct {
	Description  string          `json:"description"`
	ClearBatches bool            `json:"clear_batches"`
	ClearScripts bool            `json:"clear_scripts"`
	Filter       FilterPayload   `json:"filter"`
	Selector     SelectorOptions `json:"selector"`
}

func (c Config) String() string {
	bts, _ := json.MarshalIndent(c, " ", " ")
	return string(bts)
}

type ScriptMetadata struct {
	SelectorPayload
	InputSize  int    `json:"input_size"`
	OutputSize int    `json:"output_size"`
	Timestamp  string `json:"timestamp"`
}

func (meta ScriptMetadata) Empty() bool {
	return meta.Timestamp == "" && meta.InputSize == 0 && meta.OutputSize == 00 &&
		meta.Options.FromBatch == "" && len(meta.Options.FeatureOpts) == 0 && meta.Options.Mode == ""
}
