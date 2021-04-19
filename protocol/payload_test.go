package protocol

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

func TestPayload1(t *testing.T) {
	var err error
	filterPLB := []byte(`{"name":"filter", "message":{"level": "info", "text":"testmeddelande"}, "output_batch":"batch1", "target_size":100000, "opts":[{"name":"word_count", "args":["7", "10"]}]}`)
	selectorPLB := []byte(`{"name":"selector", "options": {"feature_opts":[{"name":"bigram_transition","target_amout":0},{"name":"bigram_top800","target_amout":3},{"name":"final_trigram","target_amout":0},{"name":"initial_bigram","target_amout":2},{"name":"word","target_amout":0},{"name":"bigram","target_amout":0}],"chunk_size":1,"chunk_decrease":0,"adjust_score_for_sentence_length":false,"target_amout":20000,"from_batch":"test_batch_1","output_script":"test_script_1","accumulated_scripts":[],"print_metadata":false,"debug":false}}`)
	payloads := [][]byte{filterPLB, selectorPLB}

	for _, plb := range payloads {

		var payload map[string]json.RawMessage
		err = json.Unmarshal(plb, &payload)
		if err != nil {
			t.Errorf("test failed: %v", err)
		}

		var name string
		err = json.Unmarshal(payload["name"], &name)
		if err != nil {
			t.Errorf("test failed: %v", err)
		}

		switch name {
		case "filter":
			var pl FilterPayload
			err = json.Unmarshal(plb, &pl)
			if err != nil {
				t.Errorf("test failed: %v", err)
			}
		case "selector":
			var pl SelectorPayload
			err = json.Unmarshal(plb, &pl)
			if err != nil {
				t.Errorf("test failed: %v", err)
			}
		default:
			t.Errorf("unknown payload type: %v", err)
		}
	}
}

func TestConfig1(t *testing.T) {

	fPL := FilterPayload{
		BatchName:  "test_batch_1",
		TargetSize: 40000,
		Opts: []FilterOpt{
			{Name: "word_count", Args: []string{"4", "25"}},
			{Name: "comma_count", Args: []string{"0", "25"}},
			{Name: "lowest_word_freq", Args: []string{"2"}},
			{Name: "exclude_chunk_re", Args: []string{"[\\p{Greek}]"}},
			{Name: "exclude_chunk_re", Args: []string{"[^a-zA-ZåäöÅÄÖéÉüÜ0-9 ,$€£@.!?/()\"':—–-]"}},
		},
	}
	sPL := SelectorOptions{
		Mode: "exhaustive", // "random"
		FeatureOpts: []SelectorFeatOpt{
			{Name: "bigram_transition"},
			{Name: "bigram_top800", TargetAmount: 3},
			{Name: "final_trigram"},
			{Name: "initial_bigram"},
			{Name: "word"},
			{Name: "bigram"},
		},
		AdjustScoreForSentenceLength: false,
		TargetSize:                   500,
		FromBatch:                    "test_batch_1",
		ScriptName:                   "test_script_1",
		//AccumulatedScripts:           []string{},
		//ContinuousPrint: false,
		PrintMetaData: true,

		// exhaustive mode
		ChunkSize:     1,
		ChunkDecrease: 0,

		// random mode
		// MinIterationsRand:            50000,
		// CutoffRand:                   100,
		Debug: false,
	}

	config := Config{Filter: fPL, Selector: sPL}

	bts, err := json.MarshalIndent(config, " ", " ")
	if err != nil {
		t.Errorf("test failed: %v", err)
	}
	fmt.Println(string(bts))

}

func TestExampleConfigFiles(t *testing.T) {
	dir := path.Join("..", "config_examples")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir : %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			bts, err := os.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				t.Errorf("failed to read %s : %v", file.Name(), err)
			}
			var config Config
			err = json.Unmarshal(bts, &config)
			if err != nil {
				t.Errorf("failed to unmarshal %s : %v", file.Name(), err)
			}

			// fmt.Println(file.Name())

			// print with indent
			// fmt.Println(config)
		}
	}
}

func TestSampleScriptConfigFiles(t *testing.T) {
	dir := path.Join("..", "sample_scripts")
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir : %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "config.json") {
			bts, err := os.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				t.Errorf("failed to read %s : %v", file.Name(), err)
			}
			var config Config
			err = json.Unmarshal(bts, &config)
			if err != nil {
				t.Errorf("failed to unmarshal %s : %v", file.Name(), err)
			}

			// fmt.Println(file.Name())

			// print with indent
			// fmt.Println(config)
		}
	}

}
