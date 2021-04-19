package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/stts-se/manuscriptor2000/dbapi"
	"github.com/stts-se/manuscriptor2000/filter"
	"github.com/stts-se/manuscriptor2000/protocol"
	"github.com/stts-se/manuscriptor2000/selection"
	"github.com/stts-se/manuscriptor2000/text"
)

const doRemoveFeats = true

func removeFeats(sents []text.Sentence) {
	if doRemoveFeats {
		for i, sent := range sents {
			sent.Feats = make(map[string]map[string]int)
			sents[i] = sent
		}
	}
}

func listFilterFeats(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	feats := filter.AvailableFeats()
	featCats, err := dbapi.ListChunkfeatCats()
	if err != nil {
		log.Printf("failed to list featcats: %v", err)
		return
	}
	seenChunkfeatCats := false

	var appendActualFeatCats = func(args0 string) string {
		return fmt.Sprintf("%s selected from: %s", args0, strings.Join(featCats, ", "))
	}

	for i, feat := range feats {
		if feat.Name == filter.ChunkFeatCats {
			feat.Args = appendActualFeatCats(feat.Args)
			feats[i] = feat
			seenChunkfeatCats = true
		}
	}
	if !seenChunkfeatCats {
		feat := filter.Feat{
			Name:    filter.ChunkFeatCats,
			Args:    appendActualFeatCats(filter.ChunkFeatCatsDocArgs),
			Desc:    filter.ChunkFeatCatsDocDesc,
			Example: filter.ChunkFeatCatsDocExample,
		}
		feats = append(feats, feat)
	}
	bts, err := json.MarshalIndent(feats, " ", " ")
	if err != nil {
		log.Printf("failed to unmarshal feats: %v", err)
		return
	}
	fmt.Println(string(bts))
}

func listSelectorFeats(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	feats := selection.AvailableFeats()
	bts, err := json.MarshalIndent(feats, " ", " ")
	if err != nil {
		log.Printf("failed to unmarshal feats: %v", err)
		return
	}
	fmt.Println(string(bts))
}

type batchOrScriptInfo struct {
	Name      string `json:"name"`
	Size      int    `json:"size,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func listBatches(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	batches, err := dbapi.ListBatchNames()
	if err != nil {
		log.Printf("failed to list batches: %v", err)
		return
	}
	if len(batches) == 0 {
		fmt.Println("No batches in db")
		return
	}
	res := []batchOrScriptInfo{}
	for _, name := range batches {
		meta, err := dbapi.GetBatchProperties(name)
		if err == nil {
			res = append(res, batchOrScriptInfo{Name: meta.BatchName, Size: meta.OutputSize, Timestamp: meta.Timestamp})
		} else {
			res = append(res, batchOrScriptInfo{Name: name})
		}
	}
	jsn, err := json.MarshalIndent(res, " ", " ")
	if err != nil {
		log.Printf("failed to marshal struct into JSON : %v", err)
		return
	}
	fmt.Println(string(jsn))
}

func listScripts(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	scripts, err := dbapi.ListScriptNames()
	if err != nil {
		log.Printf("failed to list scripts: %v", err)
		return
	}
	if len(scripts) == 0 {
		fmt.Println("No scripts in db")
		return
	}
	res := []batchOrScriptInfo{}
	for _, name := range scripts {
		meta, err := dbapi.GetScriptProperties(name)
		if err == nil {
			res = append(res, batchOrScriptInfo{Name: meta.Options.ScriptName, Size: meta.OutputSize, Timestamp: meta.Timestamp})
		} else {
			res = append(res, batchOrScriptInfo{Name: name})
		}
	}
	jsn, err := json.MarshalIndent(res, " ", " ")
	if err != nil {
		log.Printf("failed to marshal struct into JSON : %v", err)
		return
	}
	fmt.Println(string(jsn))
}

func blockSents(cmd string, args []string) {
	if len(args) == 0 {
		log.Fatalf("Args required for cmd %s", cmd)
	}
	ids := []int64{}
	for _, s := range args {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Fatalf("Failed to parse arg to string: %s", s)
		}
		ids = append(ids, i)
	}
	err := dbapi.BlockSentIDs(ids...)
	if err != nil {
		log.Printf("failed to list scripts: %v", err)
		return
	}
	fmt.Fprintf(os.Stderr, "Blocked sent ids %v\n", ids)
}

type ScriptSet struct {
	Printed string           `json:"printed"` // timestamp
	DBName  string           `json:"db_name"`
	Stats   map[string]int64 `json:"stats"`
	Scripts []Script         `json:"scripts"`
}

type Script struct {
	BatchMetadata  protocol.BatchMetadata  `json:"batch_metadata"`
	ScriptMetadata protocol.ScriptMetadata `json:"script_metadata"`
	Sentences      []text.Sentence         `json:"sentences"`
}

type ScriptWOMeta struct {
	ScriptMetadata ScriptMetaWOMeta `json:"script_metadata"`
	Sentences      []text.Sentence  `json:"sentences"`
}

type ScriptSetWOMeta struct {
	Printed string           `json:"printed"` // timestamp
	DBName  string           `json:"db_name"`
	Stats   map[string]int64 `json:"stats"`
	Scripts []ScriptWOMeta   `json:"scripts"`
}

type ScriptMetaWOMeta struct {
	OutputSize int                   `json:"output_size"`
	Options    SelectorOptionsWOMeta `json:"options"`
}

type SelectorOptionsWOMeta struct {
	ScriptName string `json:"script_name"`
}

func contains(slice []string, s string) bool {
	for _, s0 := range slice {
		if s == s0 {
			return true
		}
	}
	return false
}

func exportScripts(cmd string, args []string, includeMetaData bool) {
	var err error
	var resWithMeta = ScriptSet{
		Stats: map[string]int64{
			"size": 0,
		},
	}
	var resWOMeta = ScriptSetWOMeta{
		Stats: map[string]int64{
			"size": 0,
		},
	}
	allScripts, err := dbapi.ListScriptNames()
	if err != nil {
		log.Fatalf("dbapi.ListScriptNames failed: %v", err)
	}

	var scripts []string
	if len(args) > 0 {
		for _, script := range args {
			if !contains(allScripts, script) {
				log.Fatalf("No such script: %s", script)
			}
		}
		scripts = args
	} else {
		scripts = allScripts
	}
	for _, scriptName := range scripts {
		sents, err := dbapi.GetScript(scriptName, 0, 0)
		if err != nil {
			log.Fatalf("dbapi.GetScript failed: %v", err)
		}
		removeFeats(sents)

		if includeMetaData {
			script := Script{Sentences: sents}

			scriptMeta, err := dbapi.GetScriptProperties(scriptName)
			if err != nil {
				log.Fatalf("dbapi.GetScriptProperties failed: %v", err)
			}
			// clean up variables not relevant for export
			//scriptMeta.Options.ContinuousPrint = false
			scriptMeta.Options.Debug = false
			scriptMeta.Options.PrintMetaData = false

			batchMeta, err := dbapi.GetBatchProperties(scriptMeta.Options.FromBatch)
			if err != nil {
				log.Fatalf("dbapi.GetBatchProperties failed: %v", err)
			}
			script.BatchMetadata = batchMeta
			script.ScriptMetadata = scriptMeta
			resWithMeta.Scripts = append(resWithMeta.Scripts, script)
			resWithMeta.DBName = dbName
			resWithMeta.Stats["size"] += int64(len(sents))
			resWithMeta.Stats["scripts"] += 1
		} else {
			script := ScriptWOMeta{Sentences: sents}

			meta := ScriptMetaWOMeta{
				OutputSize: len(sents),
			}
			meta.Options = SelectorOptionsWOMeta{
				ScriptName: scriptName,
			}
			script.ScriptMetadata = meta
			resWOMeta.Scripts = append(resWOMeta.Scripts, script)
			resWOMeta.DBName = dbName
			resWOMeta.Stats["size"] += int64(len(sents))
			resWOMeta.Stats["scripts"] += 1
		}
	}
	var res interface{}
	if includeMetaData {
		resWithMeta.Printed = time.Now().Format("2006-01-02 15:04:05")
		res = resWithMeta
	} else {
		resWOMeta.Printed = time.Now().Format("2006-01-02 15:04:05")
		res = resWOMeta
	}
	bts, err := json.MarshalIndent(res, " ", " ")
	if err != nil {
		log.Fatalf("json.Marshal failed: %v", err)
	}
	fmt.Println(string(bts))
}

func exportScriptMetadata(cmd string, args []string) {
	if len(args) < 1 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	for _, scriptName := range args {
		metadata, err := dbapi.GetScriptProperties(scriptName)
		if err != nil {
			log.Fatalf("Failed to fetch script metadata %v: %v", scriptName, err)
		}
		mBts, err := json.MarshalIndent(metadata, " ", " ")
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %v", err)
		}
		fmt.Println(string(mBts))
	}
}

type BatchSet struct {
	Printed string           `json:"printed"` // timestamp
	DBName  string           `json:"db_name"`
	Stats   map[string]int64 `json:"stats"`
	Batches []Batch          `json:"batches"`
}

type Batch struct {
	Metadata  protocol.BatchMetadata `json:"metadata"`
	Sentences []text.Sentence        `json:"sentences"`
}

func exportBatches(cmd string, args []string) {
	var err error
	var res = BatchSet{
		Stats: map[string]int64{
			"size": 0,
		},
	}
	allBatches, err := dbapi.ListBatchNames()
	if err != nil {
		log.Fatalf("dbapi.ListBatchNames failed: %v", err)
	}
	var batches []string
	if len(args) > 0 {
		for _, batch := range args {
			if !contains(allBatches, batch) {
				log.Fatalf("No such batch: %s", batch)
			}
		}
		batches = args
	} else {
		batches = allBatches
	}
	for _, batchName := range batches {
		sents, err := dbapi.GetBatch(batchName, 0, 0) //opts.PageNumber, opts.PageSize)
		if err != nil {
			log.Fatalf("Failed to fetch batch %v: %v", batchName, err)
		}
		removeFeats(sents)
		batch := Batch{Sentences: sents}
		meta, err := dbapi.GetBatchProperties(batchName)
		if err != nil {
			log.Fatalf("dbapi.GetBatchProperties failed: %v", err)
		}
		batch.Metadata = meta
		res.Batches = append(res.Batches, batch)
		res.DBName = dbName
		res.Stats["size"] += int64(len(sents))
		res.Stats["batches"] += 1
	}
	res.Printed = time.Now().Format("2006-01-02 15:04:05")
	bts, err := json.MarshalIndent(res, " ", " ")
	if err != nil {
		log.Fatalf("json.Marshal failed: %v", err)
	}
	fmt.Println(string(bts))

}

func exportBatchMetadata(cmd string, args []string) {
	if len(args) < 1 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	for _, batchName := range args {
		metadata, err := dbapi.GetBatchProperties(batchName)
		if err != nil {
			log.Fatalf("Failed to fetch batch metadata %v: %v", batchName, err)
		}
		mBts, err := json.MarshalIndent(metadata, " ", " ")
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %v", err)
		}
		fmt.Println(string(mBts))
	}
}

func listBlockedSents(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	sents, err := dbapi.ListBlockedSents()
	if err != nil {
		log.Fatalf("Failed to list blocked sents: %v", err)
	}
	removeFeats(sents)
	bts, err := json.MarshalIndent(sents, " ", " ")
	if err != nil {
		log.Fatalf("Failed to marshal blocked sents: %v", err)
	}
	fmt.Println(string(bts))
}

func stats(cmd string, args []string) {
	if len(args) != 0 {
		log.Fatalf("Invalid args for cmd %s: %v", cmd, args)
	}
	stats, err := dbapi.GetStats()
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	bts, err := json.MarshalIndent(stats, " ", " ")
	if err != nil {
		log.Fatalf("Failed to marshal stats: %v", err)
	}
	fmt.Println(string(bts))
}

func runCmd(cmd string) error {
	switch cmd {
	case cHelp:
		printUsage()
	case cScriptGen:
		scriptGen(cmd, os.Args[3:])
	case cListFilterFeats:
		listFilterFeats(cmd, os.Args[3:])
	case cListSelectorFeats:
		listSelectorFeats(cmd, os.Args[3:])
	case cListBatches:
		listBatches(cmd, os.Args[3:])
	case cListScripts:
		listScripts(cmd, os.Args[3:])
	case cBlockSents:
		blockSents(cmd, os.Args[3:])
	case cListBlocked:
		listBlockedSents(cmd, os.Args[3:])
	case cExportScript:
		exportScripts(cmd, os.Args[3:], true)
	case cExportScriptWithoutMetadata:
		exportScripts(cmd, os.Args[3:], false)
	case cExportScriptMetadata:
		exportScriptMetadata(cmd, os.Args[3:])
	case cExportBatch:
		exportBatches(cmd, os.Args[3:])
	case cExportBatchMetadata:
		exportBatchMetadata(cmd, os.Args[3:])
	case cStats:
		stats(cmd, os.Args[3:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s.", cmd)
		possible := []string{}
		for _, avail := range availableCmds {
			if strings.HasPrefix(avail, cmd) {
				possible = append(possible, avail)
			}
		}
		if len(possible) == 1 {
			suggested := possible[0]
			fmt.Fprintf(os.Stderr, " Did you mean %s? [Y/n] ", suggested)
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if len(strings.TrimSpace(text)) == 0 || strings.HasPrefix(strings.ToLower(text), "y") {
				fmt.Fprintf(os.Stderr, "Running %s\n", suggested)
				runCmd(suggested)
				return nil
			}
		} else if len(possible) > 1 {
			possibleWithIndex := []string{}
			for i, pCmd := range possible {
				possibleWithIndex = append(possibleWithIndex, fmt.Sprintf("- %s [%v]", pCmd, i+1))
			}
			fmt.Fprintf(os.Stderr, " Did you mean any of these?\n%s\nChoose [1-%v] ", strings.Join(possibleWithIndex, "\n"), len(possibleWithIndex))
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			if i, err := strconv.Atoi(strings.TrimSpace(text)); err == nil && i > 0 && i <= len(possible) {
				selected := possible[i-1]
				fmt.Fprintf(os.Stderr, "Running %s\n", selected)
				runCmd(selected)
				return nil
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
		printUsage()
		os.Exit(1)
	}
	return nil
}

var terminalStderr = terminal.IsTerminal(int(os.Stderr.Fd()))
var terminalAccuPrefix = ""

var dbName string

var progName = "scripttool"

type cmd struct {
	name string
	args []string
	desc string
}

// available commands
const (
	cHelp                        = "help"
	cScriptGen                   = "scriptgen"
	cListFilterFeats             = "list_filter_feats"
	cListSelectorFeats           = "list_selector_feats"
	cListBatches                 = "list_batches"
	cListScripts                 = "list_scripts"
	cBlockSents                  = "block_sents"
	cListBlocked                 = "list_blocked_sents"
	cExportBatch                 = "export_batch"
	cExportScript                = "export_script"
	cExportScriptWithoutMetadata = "export_script_wo_metadata"
	cExportBatchMetadata         = "export_batch_metadata"
	cExportScriptMetadata        = "export_script_metadata"
	cStats                       = "stats"
)

var availableCmds = []string{
	cHelp,
	cScriptGen,
	cListFilterFeats,
	cListSelectorFeats,
	cListBatches,
	cListScripts,
	cBlockSents,
	cListBlocked,
	cExportBatch,
	cExportScript,
	cExportScriptWithoutMetadata,
	cStats,
}

var usage = []cmd{
	{name: cHelp, desc: "print help and exit"},

	{name: cScriptGen, args: []string{"config file"}, desc: "run batch filtering and/or script generation as specified in the input config file\nsample config files can be found in the config folder"},

	{name: cListFilterFeats, desc: "list available filter features"},
	{name: cListSelectorFeats, desc: "list available script features"},

	{name: cListBatches, desc: "list existing batches"},
	{name: cListScripts, desc: "list existing scripts"},

	{name: cBlockSents, args: []string{"ids"}, desc: "block specified sentence ids"},
	{name: cListBlocked, desc: "list blocked sentences"},

	{name: cExportBatch, args: []string{"batch names"}, desc: "export named batches (leave empty to export all batches)"},
	{name: cExportScript, args: []string{"script names"}, desc: "export named scripts (leave empty to export all scripts)"},

	{name: cExportScriptWithoutMetadata, args: []string{"script names"}, desc: "export named scripts without metadata (leave empty to export all batches)"},

	{name: cExportBatchMetadata, args: []string{"batch names"}, desc: "export metadata for named batches"},
	{name: cExportScriptMetadata, args: []string{"script names"}, desc: "export metadata for named scripts"},

	{name: cStats, desc: "print db statistics"},
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "USAGE: %s <Sqlite3 manuscript db> <command> <args>\n", progName)
	fmt.Fprintf(os.Stderr, "Available commands:\n")
	for _, c := range usage {
		var argsS string
		if len(c.args) == 0 {
			argsS = "none"
		} else {
			argsS = strings.Join(c.args, ",")
		}
		desc := strings.Split(c.desc, "\n")
		for i, d := range desc {
			name := c.name
			if i > 0 {
				name = ""
			}
			fmt.Fprintf(os.Stderr, " %-22s %s\n", name, d)
		}
		fmt.Fprintf(os.Stderr, "%24sargs: %s\n", "", argsS)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if terminalStderr {
		terminalAccuPrefix = "\r"
	}

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(0)
	}

	dbFile := os.Args[1]
	err := dbapi.Open(dbFile)
	if err != nil {
		log.Fatalf("Couldn't open db from file %s : %v", dbFile, err)
	}
	dbName = path.Base(dbFile)
	dbName = strings.TrimSuffix(dbName, path.Ext(dbName))

	cmd := os.Args[2]

	err = runCmd(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't run cmd %s: %v", cmd, err)
	}

	err = dbapi.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't close db: %v", err)
	}

}
