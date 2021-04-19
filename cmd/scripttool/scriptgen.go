package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
	"github.com/stts-se/wikispeech-manuscriptor/filter"
	"github.com/stts-se/wikispeech-manuscriptor/protocol"
	"github.com/stts-se/wikispeech-manuscriptor/selection"
)

func createScript(config protocol.Config) {
	opts := config.Selector
	err := selection.Validate(opts)
	if err != nil {
		log.Fatalf("Invalid selector option(s): %v", err)
	}

	selector, err := selection.NewSelector(opts)
	if err != nil {
		log.Fatalf("NewSelector failed: %v", err)
	}

	err = selector.Init()
	if err != nil {
		log.Fatalf("Init failed: %v", err)
	}

	if len(selector.Corpus) == 0 {
		log.Fatalf("Cannot run selection on empty cache!")
	}

	fmt.Fprintf(os.Stderr, "[scripttool] Selecting up to %d sentences from batch %s into script %s\n", opts.TargetSize, opts.FromBatch, opts.ScriptName)

	// if selector.Options.ContinuousPrint {
	// 	fmt.Printf("%s\n", "=== OUTPUT SCRIPT ===")
	// }

	for {
		info, ok := selector.SelectNext()
		if ok {
			if /*!options.Quiet &&*/ terminalStderr {
				fmt.Fprintf(os.Stderr, "\r% 100s", " ")
				fmt.Fprintf(os.Stderr, "\r[scripttool] Selected  : %d/%d (%d left in corpus)", len(selector.Selection), opts.TargetSize, len(selector.Corpus))
			}
			if selector.TargetReached() {
				if info == "" {
					info = "target reached"
				} else {
					info = fmt.Sprintf("target reached, %s", info)
				}
				if terminalStderr {
					fmt.Fprintf(os.Stderr, "\r% 100s", " ")
					fmt.Fprintf(os.Stderr, "\r[scripttool] Selection completed at %d sents (%s)\n", len(selector.Selection), info)
				} else {
					fmt.Fprintf(os.Stderr, "[scripttool] Selection completed at %d sents (%s)\n", len(selector.Selection), info)
				}
				break
			}
		} else {
			if terminalStderr {
				fmt.Fprintf(os.Stderr, "\r% 100s", " ")
				fmt.Fprintf(os.Stderr, "\r[scripttool] Selection completed at %d sents (%s)\n", len(selector.Selection), info)
			} else {
				fmt.Fprintf(os.Stderr, "[scripttool] Selection completed at %d sents (%s)\n", len(selector.Selection), info)
			}
			break
		}
	}

	fmt.Fprintf(os.Stderr, "[scripttool] Saving selection into script %s in db ... ", selector.Options.ScriptName)
	selectorMetadata := protocol.ScriptMetadata{SelectorPayload: protocol.SelectorPayload{Options: selector.Options}}

	// save input batch size in script metadata
	actualBatchSize, err := dbapi.BatchSize(opts.FromBatch)
	if err != nil {
		log.Fatalf("dbapi.BatchSize failed : %v", err)
	}
	selectorMetadata.InputSize = int(actualBatchSize)
	selectorMetadata.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	n, err := selector.WriteScriptToDB(selectorMetadata)
	if err != nil {
		log.Fatalf("Save selection failed: %v", err)
	}
	selectorMetadata.OutputSize = n
	fmt.Fprintf(os.Stderr, "done (saved %d sentences)\n", n)

	fmt.Fprintf(os.Stderr, "\n=== SCRIPT COMPLETED ===\n")
	fmt.Fprintf(os.Stderr, "Name: %s\n", opts.ScriptName)
	fmt.Fprintf(os.Stderr, "Target size: %v\n", opts.TargetSize)
	fmt.Fprintf(os.Stderr, "Output size: %v\n", selectorMetadata.OutputSize)
	fmt.Fprintf(os.Stderr, "Timestamp: %s\n", selectorMetadata.Timestamp)
}

func createBatch(config protocol.Config) {
	batchMetadata := protocol.BatchMetadata{FilterPayload: config.Filter}
	filterQueryBuilder, err := filter.NewQueryBuilder(config.Filter)
	if err != nil {
		log.Fatalf("Couldn't create query builder : %v", err)
	}

	fmt.Fprintf(os.Stderr, "[scripttool] Filtering up to %v sents into batch %s\n", config.Filter.TargetSize, config.Filter.BatchName)
	n, err := filter.ExecQuery(filterQueryBuilder)
	if err != nil {
		log.Fatalf("Couldn't exec query : %v", err)
	}

	batchMetadata.OutputSize = int(n)
	batchMetadata.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	fmt.Fprintf(os.Stderr, "\n=== BATCH COMPLETED ===\n")
	fmt.Fprintf(os.Stderr, "Name: %s\n", batchMetadata.FilterPayload.BatchName)
	fmt.Fprintf(os.Stderr, "Target size: %v\n", batchMetadata.TargetSize)
	fmt.Fprintf(os.Stderr, "Output size: %v\n", batchMetadata.OutputSize)
	fmt.Fprintf(os.Stderr, "Timestamp: %s\n", batchMetadata.Timestamp)
	fmt.Fprintf(os.Stderr, "\n")

	pBytes, err := json.Marshal(batchMetadata)
	if err != nil {
		log.Fatalf("failed to marshal filter metadata : %v", err)
	}
	err = dbapi.SetBatchProperties(config.Filter.BatchName, pBytes)
	if err != nil {
		log.Fatalf("failed to save batch properties : %v", err)
	}

	if n == 0 {
		log.Fatalf("Output batch was empty!")
	}
}

func scriptGen(cmd string, args []string) {
	if len(args) != 1 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	configFile := args[0]
	var config protocol.Config
	bts, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file %s : %v", configFile, err)
	}
	err = json.Unmarshal(bts, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file %s : %v", configFile, err)
	}

	if config.Filter.TargetSize == 0 && !config.Filter.Empty() {
		fmt.Fprintf(os.Stderr, "If filter config is non-empty, target size must be specified\n")
		os.Exit(1)
	}

	if config.Filter.TargetSize > 0 && config.Filter.TargetSize < config.Selector.TargetSize {
		fmt.Fprintf(os.Stderr, "Batch target size cannot be higher than script target size (%v vs %v)\n", config.Filter.TargetSize, config.Selector.TargetSize)
		os.Exit(1)
	}

	if config.ClearBatches {
		fmt.Fprintf(os.Stderr, "[scripttool] Clearing batch %s... ", config.Filter.BatchName)
		err = dbapi.DeleteBatches(config.Filter.BatchName)
		if err != nil {
			log.Fatalf("DeleteBatches failed: %v", err)
		}
		fmt.Fprintf(os.Stderr, "done\n")
	}
	if config.ClearScripts {
		fmt.Fprintf(os.Stderr, "[scripttool] Clearing script %s... ", config.Selector.ScriptName)
		err = dbapi.DeleteScripts(config.Selector.ScriptName)
		if err != nil {
			log.Fatalf("DeleteScripts failed: %v", err)
		}
		fmt.Fprintf(os.Stderr, "done\n")
	}

	if !config.Filter.Empty() {
		createBatch(config)
	}

	if !config.Selector.Empty() {
		createScript(config)
	}
}
