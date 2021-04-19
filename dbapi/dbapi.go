package dbapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/mattn/go-sqlite3"

	"github.com/stts-se/manuscriptor2000/protocol"
	"github.com/stts-se/manuscriptor2000/text"
)

var db *sql.DB

var remem = struct {
	sync.Mutex
	re map[string]*regexp.Regexp
}{
	re: make(map[string]*regexp.Regexp),
}

var regexMem = func(re, s string) (bool, error) {

	remem.Lock()
	defer remem.Unlock()
	if r, ok := remem.re[re]; ok {
		return r.MatchString(s), nil
	}

	r, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}
	remem.re[re] = r
	return r.MatchString(s), nil
}

// Sqlite3WithRegex registers an Sqlite3 driver with regexp support. (Unfortunately quite slow regexp matching)
func Sqlite3WithRegex() {
	// regex := func(re, s string) (bool, error) {
	// 	//return regexp.MatchString(re, s)
	// 	return regexp.MatchString(re, s)
	// }
	sql.Register("sqlite3_with_regexp",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regexMem, true)
			},
		})
}

func openDB(dbPath string, createIfNotExists bool, pragmas ...string) error {
	Sqlite3WithRegex()

	if !createIfNotExists {
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return fmt.Errorf("file doesn't exist: '%s'", dbPath)
		}
	}

	db0, err := sql.Open("sqlite3_with_regexp", dbPath)

	if err != nil {
		return fmt.Errorf("failed to open db file '%s' : %v", dbPath, err)
	}

	_, err = db0.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("failed to run PRAGMA foreign_keys = ON : %v", err)
	}

	for _, p := range pragmas {
		_, err = db0.Exec(p + ";")
		if err != nil {
			return fmt.Errorf("failed to run '%s' : %v", p, err)
		}
	}

	db = db0
	return nil
}

func CreateDB(dbPath string, schemaFile string) error {
	err := openDB(dbPath, true)
	if err != nil {
		return fmt.Errorf("failed to open db : %v", err)
	}
	bts, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("failed to read %s : %v", schemaFile, err)
	}

	err = ExecSchema(string(bts))
	if err != nil {
		log.Fatalf("failed to exec schema : %v", err)
	}
	return nil
}

func Open(dbPath string, pragmas ...string) error {
	err := openDB(dbPath, false, pragmas...)
	if err != nil {
		return fmt.Errorf("failed to open db : %v", err)
	}
	return nil
}

func Close() error {
	return db.Close()
}

func Begin() (*sql.Tx, error) {
	return db.Begin()
}

func ForeignKeysOn() error {
	_, err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		return fmt.Errorf("failed to turn on foreign_keys : %v", err)
	}

	return nil
}

func CacheSize(n int) error {
	_, err := db.Exec(fmt.Sprintf("PRAGMA cache_size = %d", n))
	if err != nil {
		return fmt.Errorf("failed to turn on foreign_keys : %v", err)
	}

	return nil
}

func ForeignKeysOff() error {
	_, err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		return fmt.Errorf("failed to turn off foreign_keys : %v", err)
	}

	return nil
}
func JournalModeOff() error {
	_, err := db.Exec("PRAGMA journal_mode = OFF")
	if err != nil {
		return fmt.Errorf("failed to turn off journal_mode : %v", err)
	}

	return nil
}

func SynchronousOff() error {
	_, err := db.Exec("PRAGMA synchronous = OFF")
	if err != nil {
		return fmt.Errorf("failed to turn off synchronous : %v", err)
	}

	return nil
}

func SynchronousOn() error {
	_, err := db.Exec("PRAGMA synchronous = ON")
	if err != nil {
		return fmt.Errorf("failed to turn on synchronous : %v", err)
	}

	return nil
}

func JournalModeMemory() error {
	_, err := db.Exec("PRAGMA journal_mode = MEMORY")
	if err != nil {
		return fmt.Errorf("failed to turn on journal_mode=MEMORY : %v", err)
	}

	return nil
}

func MaxRowID(tableName string) (int64, error) {
	var res sql.NullInt64

	// TODO Why doesn't it work with '?' placeholder syntax
	// db.QueryRow(`SELECT MAX(_ROWID_) FROM ? LIMIT 1`, tableName) ???

	err := db.QueryRow(`SELECT MAX(_ROWID_) FROM ` + tableName + ` LIMIT 1`).Scan(&res)
	if err != nil {
		return res.Int64, fmt.Errorf("failed to get MAX _ROWID_ for '%s' : %v", tableName, err)
	}

	return res.Int64, nil
}

type Stats struct {
	Chunks      int64          `json:"chunks"`
	Sources     int64          `json:"sources"`
	ChunkFeats  int64          `json:"chunk_feats"`
	WordForms   int64          `json:"word_forms"`
	MaxWordFreq int64          `json:"max_wordfreq"`
	Batches     map[string]int `json:"batches"`
	Scripts     map[string]int `json:"scripts"`
}

func GetStats() (Stats, error) {
	res := Stats{}

	// TODO: MaxRowID is not secure
	chnks, err := MaxRowID("chunk")
	if err != nil {
		return res, fmt.Errorf("failed MaxRowID for chunk : %v", err)
	}

	res.Chunks = chnks

	srcs, err := MaxRowID("source")
	if err != nil {
		return res, fmt.Errorf("failed MaxRowID for source : %v", err)
	}

	res.Sources = srcs

	cfeats, err := MaxRowID("chunkfeat")
	if err != nil {
		return res, fmt.Errorf("failed MaxRowID for chunkfeat : %v", err)
	}

	res.ChunkFeats = cfeats

	var wds sql.NullInt64

	err = db.QueryRow("SELECT COUNT(*) FROM chunkfeat WHERE chunkfeat.name = ?", text.FeatWord).Scan(&wds)
	if err != nil {
		return res, fmt.Errorf("failed count words from chunkfeat : %v", err)
	}

	res.WordForms = wds.Int64

	var maxFreq sql.NullInt64
	err = db.QueryRow("SELECT wordfreq.freq FROM wordfreq WHERE id = 1").Scan(&maxFreq)
	if err != nil {
		return res, fmt.Errorf("failed to max word freq : %v", err)
	}

	res.MaxWordFreq = maxFreq.Int64

	rows, err := db.Query("SELECT name, count(name) FROM batch group by name")
	if err != nil {
		return res, fmt.Errorf("failed to read batches : %v", err)
	}
	batches := make(map[string]int)
	for rows.Next() {
		var b string
		var c int
		rows.Scan(&b, &c)
		batches[b] = c
	}
	res.Batches = batches

	rows, err = db.Query("SELECT name, count(name) FROM script group by name")
	if err != nil {
		return res, fmt.Errorf("failed to read scripts : %v", err)
	}
	scripts := make(map[string]int)
	for rows.Next() {
		var s string
		var c int
		rows.Scan(&s, &c)
		scripts[s] = c
	}
	res.Scripts = scripts

	return res, nil
}

func ExecSchema(schema string) error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to exec schema : %v", err)
	}
	return nil
}

func ExecQuery(query string, args []interface{}) (*sql.Rows, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return rows, fmt.Errorf("failed to exec schema : %v", err)
	}
	return rows, nil
}

func BlockSentIDs(ids ...int64) error {
	tx, err := Begin()
	if err != nil {
		log.Fatalf("BlockSentIDs failed to begin db transaction : %v", err)
	}
	qs := []string{}
	args := []interface{}{text.BlockBatch}
	for _, sent := range ids {
		qs = append(qs, "?")
		args = append(args, sent)
	}
	query := fmt.Sprintf("INSERT INTO batch (chunk_id, name) SELECT chunk.id, ? FROM chunk WHERE chunk.id IN (%s)", strings.Join(qs, ","))

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("BlockSentIDs failed to insert : %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit transaction : %v", err)
	}
	return nil
}

func ListBlockedSents() ([]text.Sentence, error) {
	tx, err := Begin()
	if err != nil {
		log.Fatalf("ListBlockedSents failed to begin db transaction : %v", err)
	}
	query := fmt.Sprintf("SELECT chunk_id FROM batch WHERE batch.name = ?")

	rows, err := tx.Query(query, text.BlockBatch)
	if err != nil {
		tx.Rollback()
		return []text.Sentence{}, fmt.Errorf("ListBlockedSents failed to query : %v", err)
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	sents, err := GetSents(ids...)
	if err != nil {
		tx.Rollback()
		return []text.Sentence{}, fmt.Errorf("ListBlockedSents failed to get sents from ids : %v", err)
	}

	if err = rows.Err(); err != nil {
		return []text.Sentence{}, fmt.Errorf("error when reading db result row : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return []text.Sentence{}, fmt.Errorf("couldn't commit transaction : %v", err)
	}
	return sents, nil
}

func BlockSents(sents ...string) error {
	tx, err := Begin()
	if err != nil {
		log.Fatalf("BlockSents failed to begin db transaction : %v", err)
	}
	qs := []string{}
	args := []interface{}{text.BlockBatch}
	for _, sent := range sents {
		qs = append(qs, "?")
		args = append(args, sent)
	}
	query := fmt.Sprintf("INSERT INTO batch (chunk_id, name) SELECT chunk.id, ? FROM chunk WHERE chunk.text IN (%s)", strings.Join(qs, ","))

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("BlockSents failed to insert : %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit transaction : %v", err)
	}
	return nil
}

// Add an article, return article id, inserted sents (with ids), and error if any
func Add(a text.Article, insertChunkFeats bool) (int64, []text.Sentence, error) {

	tx, err := Begin()
	if err != nil {
		log.Fatalf("dbapi.Add failed to begin db transaction : %v", err)
	}

	sents := []text.Sentence{}

	sID, err := InsertSourceTx(tx, a.URL)
	if err != nil {
		return sID, sents, fmt.Errorf("dbapi.Add failed InsertSourceTx: %v", err)
	}

	s := SourceFeaturesExtractor{}
	feats := s.Extract(a)

	err = InsertSourceFeatsTx(tx, sID, feats)
	if err != nil {
		return sID, sents, fmt.Errorf("dbapi.Add failed InsertSourceFeatsTx: %v", err)
	}

	source := a.URL
	for _, p := range a.Paragraphs {
		for _, s := range p.Sentences {
			_, cID, newChunk, err := InsertChunkTx(tx, source, s.Text)
			if err != nil {
				tx.Rollback()
				return sID, sents, fmt.Errorf("dbapi.Add failed to insert chunk into DB : %v", err)
			}

			s.ID = cID
			if newChunk {
				sents = append(sents, s)
			}
			if insertChunkFeats && newChunk {
				err = InsertChunkFeatsTx(tx, cID, s.Feats)
				if err != nil {
					tx.Rollback()
					return sID, sents, fmt.Errorf("dbapi.Add failed to insert chunkfeats into DB : %v", err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return sID, sents, fmt.Errorf("dbapi.Add failed to commit transaction : %v", err)
	}

	return sID, sents, nil
}

// PopulateWordFreqTable() adds word frequencies (number of chunks a word occurs in) and should be run _after_ the initail corpus data have been added to the db
func PopulateWordFreqTable() error {

	//log.Println("Started generating word frequency table")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("PopulateWordFreqTable failed to start db transaction : %v", err)
	}

	_, err = tx.Exec("DELETE FROM wordfreq;")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PopulateWordFreqTable failed to empty wordfreq table : %v", err)
	}

	_, err = tx.Exec(`INSERT INTO wordfreq (chunkfeat_id, freq) SELECT chunk_chunkfeat.chunkfeat_id, COUNT(*) FROM chunk_chunkfeat, chunkfeat WHERE chunkfeat.name = ? AND chunkfeat.id = chunk_chunkfeat.chunkfeat_id GROUP BY chunk_chunkfeat.chunkfeat_id ORDER BY COUNT(chunk_chunkfeat.chunk_id) DESC;`, text.FeatWord)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PopulateWordFreqTable failed to generate wordfreq table : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("PopulateWordFreqTable failed to commit db transaction : %v", err)
	}

	return nil
}

// InsertLeastWordFreqForChunk adds the frequency of the least frequent word of a chunk into the chunk_chunkfeat table, by the relation chunkfeat.name = 'count' and chunfeat.value = 'lowest_word_freq'.
// MUST be called *after* PopulateWordFreqTable()
func InsertLowestWordFreqForChunk() error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("InsertLowestWordFreqForChunk failed to start db transaction : %v", err)
	}

	// get id of featur/value count/lowest_word_freq
	fName := text.FeatCount
	vName := text.FeatValLowestWordFreq // "lowest_word_freq"
	var fID sql.NullInt64
	err = tx.QueryRow("SELECT id FROM chunkfeat WHERE name = ? AND value = ?", fName, vName).Scan(&fID)
	featID := fID.Int64

	switch {
	case err == sql.ErrNoRows:

		execRes, err := tx.Exec("INSERT INTO chunkfeat(name, value) VALUES(?, ?)", fName, vName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert chunkfeat %s %s", fName, vName)
		}
		featID, err = execRes.LastInsertId()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed LastInserId() : %v", err)
		}

	case err == nil && featID != 0:
		// TODO: TEST
		// already existing count/lowest_word_count feature: delete previuos values
		_, err := tx.Exec("DELETE FROM chunk_chunkfeat WHERE chunk_chunkfeat.chunkfeat_id = ?", featID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed DELETE %s/%s feature value: %v", text.FeatCount, text.FeatValLowestWordFreq, err)
		}
	case err != nil:
		tx.Rollback()
		return fmt.Errorf("failed QueryRow : %v", err)

	}

	featIDstr := strconv.FormatInt(featID, 10)

	_, err = tx.Exec(`INSERT INTO chunk_chunkfeat (chunk_id, freq, chunkfeat_id) SELECT chunk.id, MIN(wordfreq.freq), ` + featIDstr + ` FROM chunk JOIN chunk_chunkfeat on chunk_chunkfeat.chunk_id = chunk.id JOIN wordfreq ON wordfreq.chunkfeat_id = chunk_chunkfeat.chunkfeat_id GROUP BY chunk.id;`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute insert statement for lowest word freq calculation : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction : %v", err)
	}
	return nil
}

// TODO make batch versions of insert funcs --- better performance if more stuff is added in single transaction?

// InsertSource adds a new source string s to the db and returns its db id. If s already exists, returns its existing id.
func InsertSource(s string) (int64, error) {
	var res int64

	s = strings.TrimSpace(s)

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	row := tx.QueryRow("SELECT id FROM source WHERE name = ?", s).Scan(&res)
	if res > 0 && row != sql.ErrNoRows {
		// Source already exists
		err = tx.Commit()
		if err != nil {
			return res, fmt.Errorf("couldn't commit transaction : %v", err)
		}
		return res, nil
	}

	execRes, err := tx.Exec("INSERT INTO source (name) VALUES (?)", s)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to execure insert statement : %v", err)
	}

	res, err = execRes.LastInsertId()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to get last insert id : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to commit transaction : %v", err)
	}

	return res, nil

}

func InsertSourceTx(tx *sql.Tx, s string) (int64, error) {
	var res int64

	s = strings.TrimSpace(s)

	// tx, err := db.Begin()
	// if err != nil {
	// 	tx.Rollback()
	// 	return res, fmt.Errorf("failed to begin transaction : %v", err)
	// }

	row := tx.QueryRow("SELECT id FROM source WHERE name = ?", s).Scan(&res)
	if res > 0 && row != sql.ErrNoRows {
		// Source already exists
		//tx.Commit()
		return res, nil
	}

	execRes, err := tx.Exec("INSERT INTO source (name) VALUES (?)", s)
	if err != nil {
		//tx.Rollback()
		return res, fmt.Errorf("failed to execure insert statement : %v", err)
	}

	res, err = execRes.LastInsertId()
	if err != nil {
		//tx.Rollback()
		return res, fmt.Errorf("failed to get last insert id : %v", err)
	}

	//err = tx.Commit()
	if err != nil {
		//tx.Rollback()
		return res, fmt.Errorf("failed to commit transaction : %v", err)
	}

	return res, nil

}

func InsertSourceFeatsTx(tx *sql.Tx, sourceID int64, feats []Feat) error {

	var err error

	for _, f := range feats {
		var id sql.NullInt64
		err = tx.QueryRow("SELECT id FROM sourcefeat WHERE name = ? AND value = ?", f.Name, f.Value).Scan(&id)
		sourceFeatID := id.Int64

		switch {
		case err == sql.ErrNoRows:

			execRes, err := tx.Exec("INSERT INTO sourcefeat(name, value)  VALUES(?, ?)", f.Name, f.Value)
			if err != nil {
				return fmt.Errorf("failed to insert sourcefeat %s %s", f.Name, f.Value)
			}
			sourceFeatID, err = execRes.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed LastInserId() : %v", err)
			}

		case err != nil:
			return fmt.Errorf("failed QueryRow : %v", err)

		}

		// Now the sourcefeat is in the db, lets add the source - sourcefeat relation

		var cID sql.NullInt64
		err = tx.QueryRow("SELECT source_id from source_sourcefeat WHERE source_id = ? AND sourcefeat_id = ?", sourceID, sourceFeatID).Scan(&cID)

		// No relation established
		if err == sql.ErrNoRows {
			_, err := tx.Exec("INSERT INTO source_sourcefeat(source_id, sourcefeat_id, freq) VALUES(?, ?, ?)", sourceID, sourceFeatID, f.Freq)
			if err != nil {
				//tx.Rollback()
				return fmt.Errorf("failed to insert source_sourcefeat relation : %v", err)
			}
		} else if err != nil {
			//tx.Rollback()
			return fmt.Errorf("Query row failure : %v", err)
		}

	}

	return nil
}

func InsertChunk(source string, chunk string) (int64, int64, error) {
	var res int64

	sourceID, err := InsertSource(source)
	if err != nil {
		return sourceID, res, fmt.Errorf("failed InsertSource : %v", err)
	}

	s := strings.TrimSpace(chunk)

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return sourceID, res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	row := tx.QueryRow("SELECT id FROM chunk WHERE text = ?", s).Scan(&res)
	if row == sql.ErrNoRows {

		execRes, err := tx.Exec("INSERT INTO chunk (text) VALUES (?)", s)
		if err != nil {
			tx.Rollback()
			return sourceID, res, fmt.Errorf("failed to execure insert statement : %v", err)
		}

		res, err = execRes.LastInsertId()
		if err != nil {
			tx.Rollback()
			return sourceID, res, fmt.Errorf("failed to get last insert id : %v", err)
		}
	} else if row != nil {
		tx.Rollback()
		return sourceID, res, fmt.Errorf("QueryRow failure : %v", err)
	}

	// Create relation
	var tmp1, tmp2 sql.NullInt64
	relationRow := tx.QueryRow("SELECT source_id, chunk_id FROM source_chunk WHERE source_id = ? AND chunk_id = ?", sourceID, res).Scan(&tmp1, &tmp2)
	if relationRow == sql.ErrNoRows /*relationRow == nil*/ {

		_, err := tx.Exec("INSERT INTO source_chunk (source_id, chunk_id) VALUES(?, ?)", sourceID, res)
		if err != nil {
			tx.Rollback()
			return sourceID, res, fmt.Errorf("failed to insert source-chunk relation : %v", err)
		}
	} else if relationRow != nil {
		tx.Rollback()
		return sourceID, res, fmt.Errorf("QueryRow error : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return sourceID, res, fmt.Errorf("failed to commit transaction : %v", err)
	}

	return sourceID, res, nil

}

// InsertChunkTx returns sourceID, chunkID, newChunk (bool), error
func InsertChunkTx(tx *sql.Tx, source string, chunk string) (int64, int64, bool, error) {
	var res int64
	var res0 sql.NullInt64
	var newChunk = false

	sourceID, err := InsertSourceTx(tx, source)
	if err != nil {
		return sourceID, res, newChunk, fmt.Errorf("failed InsertSource : %v", err)
	}

	s := strings.TrimSpace(chunk)

	row := tx.QueryRow("SELECT id FROM chunk WHERE text = ?", s).Scan(&res0)
	res = res0.Int64
	if row == sql.ErrNoRows {
		newChunk = true

		execRes, err := tx.Exec("INSERT INTO chunk (text) VALUES (?)", s)
		if err != nil {
			//tx.Rollback()
			return sourceID, res0.Int64, newChunk, fmt.Errorf("failed to execure insert statement : %v", err)
		}

		res, err = execRes.LastInsertId()
		if err != nil {
			//tx.Rollback()
			return sourceID, res, newChunk, fmt.Errorf("failed to get last insert id : %v", err)
		}
	} else if row != nil {
		//tx.Rollback()
		return sourceID, res0.Int64, newChunk, fmt.Errorf("QueryRow failure : %v", err)
	}

	// Create relation
	var tmp1, tmp2 sql.NullInt64
	relationRow := tx.QueryRow("SELECT source_id, chunk_id FROM source_chunk WHERE source_id = ? AND chunk_id = ?", sourceID, res).Scan(&tmp1, &tmp2)
	if relationRow == sql.ErrNoRows /*relationRow == nil*/ {

		_, err := tx.Exec("INSERT INTO source_chunk (source_id, chunk_id) VALUES(?, ?)", sourceID, res)
		if err != nil {
			//tx.Rollback()
			return sourceID, res, newChunk, fmt.Errorf("failed to insert source-chunk relation : %v", err)
		}
	} else if relationRow != nil {
		//tx.Rollback()
		return sourceID, res, newChunk, fmt.Errorf("QueryRow error : %v", err)
	}

	//err = tx.Commit()
	if err != nil {
		//tx.Rollback()
		return sourceID, res, newChunk, fmt.Errorf("failed to commit transaction : %v", err)
	}

	return sourceID, res, newChunk, nil

}

type Feat struct {
	Name  string
	Value string
	Freq  int
}

func MostFrequentWords(limit int) ([]string, error) {

	var res []string

	rows, err := db.Query(`SELECT chunkfeat.value FROM wordfreq, chunkfeat WHERE wordfreq.chunkfeat_id = chunkfeat.id AND chunkfeat.name = ? ORDER BY wordfreq.freq DESC LIMIT ?`, text.FeatWord, limit)
	if err != nil {
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var w string
		rows.Scan(&w)
		res = append(res, w)
	}

	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}

	return res, nil
}

func ListChunkfeatCats() ([]string, error) {
	var res = []string{}

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	rows, err := tx.Query(`SELECT DISTINCT name FROM chunkfeatcat`)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}
	tx.Commit()
	return res, nil
}

func ListBatches() ([]protocol.BatchMetadata, error) {
	var res []protocol.BatchMetadata

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	rows, err := tx.Query(`SELECT DISTINCT name FROM batch where name <> ?`, "blocked")
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		props, err := GetBatchPropertiesTx(tx, name)
		if err != nil {
			tx.Rollback()
			return res, err
		}
		res = append(res, props)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}
	tx.Commit()
	return res, nil
}

func ListBatchNames() ([]string, error) {
	var res []string

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	rows, err := tx.Query(`SELECT DISTINCT name FROM batch`)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}
	tx.Commit()
	return res, nil
}

func ListScripts() ([]protocol.ScriptMetadata, error) {
	var res []protocol.ScriptMetadata

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	rows, err := tx.Query(`SELECT DISTINCT name FROM script`)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		props, err := GetScriptPropertiesTx(tx, name)
		if err != nil {
			tx.Rollback()
			return res, err
		}
		res = append(res, props)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}
	tx.Commit()
	return res, nil
}

func ListScriptNames() ([]string, error) {
	var res []string

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	rows, err := tx.Query(`SELECT DISTINCT name FROM script`)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to query db : %v", err)
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		res = append(res, name)
	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error when reading db result row : %v", err)
	}
	tx.Commit()
	return res, nil
}

func DeleteBatches(batches ...interface{}) error {
	var qs []string
	for i := 0; i < len(batches); i++ {
		qs = append(qs, "?")
	}
	query := fmt.Sprintf("DELETE FROM batch WHERE name IN ( %s )", strings.Join(qs, ", "))
	_, err := db.Exec(query, batches...)
	if err != nil {
		return fmt.Errorf("failed to delete batches '%v' : %v", batches, err)
	}
	query = fmt.Sprintf("DELETE FROM batch_properties WHERE name IN (%s)", strings.Join(qs, ", "))
	_, err = db.Exec(query, batches...)
	if err != nil {
		return fmt.Errorf("failed to delete batch_properties '%v' : %v", batches, err)
	}

	return nil
}

func DeleteScripts(scripts ...interface{}) error {
	var qs []string
	for i := 0; i < len(scripts); i++ {
		qs = append(qs, "?")
	}
	query := fmt.Sprintf("DELETE FROM script WHERE name IN (%s)", strings.Join(qs, ", "))
	_, err := db.Exec(query, scripts...)
	if err != nil {
		return fmt.Errorf("failed to delete scripts '%v' : %v", scripts, err)
	}

	query = fmt.Sprintf("DELETE FROM script_properties WHERE name IN (%s)", strings.Join(qs, ", "))
	_, err = db.Exec(query, scripts...)
	if err != nil {
		return fmt.Errorf("failed to delete script_properties '%v' : %v", scripts, err)
	}

	return nil
}

func BatchSize(batches ...interface{}) (int, error) {
	var qs []string
	for i := 0; i < len(batches); i++ {
		qs = append(qs, "?")
	}
	query := fmt.Sprintf("SELECT COUNT(*) FROM batch WHERE name IN ( %s )", strings.Join(qs, ", "))
	tx, err := Begin()
	if err != nil {
		return 0, fmt.Errorf("couldn't create db connection : %v", err)
	}
	var count int
	err = tx.QueryRow(query, batches...).Scan(&count)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to compute batch size '%v' : %v", batches, err)
	}
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	if count >= 0 {
		return count, nil
	}
	return 0, fmt.Errorf("failed to compute batch size '%v'", batches)
}

func GetSentsInBatches(batchSize int, ids ...int64) ([]text.Sentence, error) {
	var res []text.Sentence

	var batch []int64
	for _, id := range ids {
		batch = append(batch, id)
		if len(batch) >= batchSize {
			snts, err := GetSents(batch...)
			if err != nil {
				return res, err
			}
			//fmt.Fprintf(os.Stderr, "dbapi.GetSentsInBatches accumulated %v\n", len(res))
			res = append(res, snts...)
			batch = []int64{}
		}
	}

	if len(batch) > 0 {
		snts, err := GetSents(batch...)
		if err != nil {
			return res, err
		}
		res = append(res, snts...)
	}
	//fmt.Fprintf(os.Stderr, "dbapi debug GetSentsInBatches accumulated %v\n", len(res))
	return res, nil
}

func GetSents(ids ...int64) ([]text.Sentence, error) {
	//fmt.Fprintf(os.Stderr, "GetSents called with %d ids\n", len(ids))
	var res []text.Sentence

	tx, err := Begin()
	if err != nil {
		return res, fmt.Errorf("failed to start transaction : %v", err)
	}

	res, err = GetSentsTx(tx, ids...)
	if err != nil {
		return res, fmt.Errorf("failed GetSentsTx: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

func GetSentsTx(tx *sql.Tx, ids ...int64) ([]text.Sentence, error) {
	var res []text.Sentence
	var err error

	tmpTableName := fmt.Sprintf("tmp_get_sents_%s", text.RandomString(10))
	tmpTableName = strings.Replace(tmpTableName, "-", "_", -1)

	_, err = tx.Exec("CREATE TEMP TABLE IF NOT EXISTS " + tmpTableName + " (id INTEGER NOT NULL UNIQUE)")
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to create tmp table '%s': %v", tmpTableName, err)
	}
	_, err = tx.Exec("CREATE INDEX index_" + tmpTableName + " ON " + tmpTableName + "(id)")
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to create index for temp table '%s': %v", tmpTableName, err)
	}

	for _, id := range ids {
		_, err := tx.Exec("INSERT INTO "+tmpTableName+" (id) VALUES (?)", id)
		if err != nil {
			tx.Rollback()
			return res, fmt.Errorf("failed to insert into tmp table '%s': %v", tmpTableName, err)
		}
	}
	//fmt.Fprintf(os.Stderr, "dbapi debug Inserted %v ids into %s\n", len(ids), tmpTableName)

	// select all chunks from tmp and turn into text.Sentence
	q := "SELECT chunk.id, chunk.text, chunkfeat.name, chunkfeat.value, chunk_chunkfeat.freq, chunkfeatcat.name, source.name FROM chunk, chunk_chunkfeat, chunkfeat, source, source_chunk LEFT JOIN chunkfeatcat ON chunkfeatcat.chunkfeat_id = chunkfeat.id WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunk_chunkfeat.chunkfeat_id = chunkfeat.id AND source_chunk.chunk_id = chunk.id AND source_chunk.source_id = source.id AND chunk.id IN (SELECT id FROM " + tmpTableName + ")"
	rows, err := tx.Query(q)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to select from chunk table : %v", err)
	}

	//fmt.Fprintf(os.Stderr, "dbapi debug Fetched lines from db\n")

	tmpRes := make(map[int64]text.Sentence)

	var currSent text.Sentence
	i := 0
	for rows.Next() {
		i++
		// if i%1000000 == 0 {
		// 	fmt.Fprintf(os.Stderr, "dbapi debug %v lines\n", i)
		// }

		var id, freq int64
		var chunk, fName, fVal, source string
		var fCat sql.NullString
		err := rows.Scan(&id, &chunk, &fName, &fVal, &freq, &fCat, &source)
		if err != nil {
			tx.Rollback()
			return res, fmt.Errorf("failed to scan rows : %v", err)
		}

		//fmt.Fprintf(os.Stderr, "%v|%v|%v|%v|%v|%v|%v\n", id, chunk, fName, fVal, freq, fCat, source)

		if s, ok := tmpRes[id]; ok {
			currSent = s
		} else {
			currSent = text.Sentence{ID: id, Text: chunk, Feats: make(map[string]map[string]int), Source: source}
			//fmt.Fprintf(os.Stderr, "New sentence %v\n", currSent.ID)
			//fmt.Fprintf(os.Stderr, "dbapi debug %v sents\n", len(tmpRes))
			tmpRes[id] = currSent
		}

		currSent.AddFeatWithFreq(fName, fVal, int(freq))
		if fCat.Valid {
			currSent.AddFeatWithFreq(fCat.String, fVal, int(freq))
		}
	}

	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("error when reading result row : %v", err)
	}

	// Empty the tmp table
	_, err = tx.Exec("DELETE FROM " + tmpTableName)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to empty tmp table '%s': %v", tmpTableName, err)
	}

	for _, id := range ids {
		s, ok := tmpRes[id]
		if !ok {
			return res, fmt.Errorf("failed to find chunk id '%d' in query resul", id)
		}
		res = append(res, s)
	}

	err = rows.Close()
	if err != nil {
		return res, fmt.Errorf("couldn't close rows : %v", err)
	}
	return res, nil
}

func GetScriptPropertiesTx(tx *sql.Tx, scriptName string) (protocol.ScriptMetadata, error) {
	var res protocol.ScriptMetadata

	var bytes []byte
	err := db.QueryRow(`SELECT properties FROM script_properties WHERE name = ?`, scriptName).Scan(&bytes)
	if err != nil {
		return res, fmt.Errorf("failed to exec query : %v", err)
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't unmarshal batch properties : %v", err)
	}
	return res, nil
}

func GetScriptProperties(scriptName string) (protocol.ScriptMetadata, error) {
	var res protocol.ScriptMetadata

	tx, err := Begin()
	if err != nil {
		return res, fmt.Errorf("failed to start transaction : %v", err)
	}

	res, err = GetScriptPropertiesTx(tx, scriptName)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("no script properties for %s: %v", scriptName, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

func SetScriptPropertiesTx(tx *sql.Tx, scriptName string, properties []byte) error {
	_, err := tx.Exec("INSERT INTO script_properties (name, properties) VALUES (?, ?)", scriptName, properties)
	if err != nil {
		return fmt.Errorf("failed to exec query : %v", err)
	}
	return nil
}

func SetScriptProperties(scriptName string, properties []byte) error {
	tx, err := Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction : %v", err)
	}

	err = SetScriptPropertiesTx(tx, scriptName, properties)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("couldn't set script properties: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return nil
}

func GetBatchPropertiesTx(tx *sql.Tx, batchName string) (protocol.BatchMetadata, error) {
	var res protocol.BatchMetadata

	var bytes []byte
	err := db.QueryRow(`SELECT properties FROM batch_properties WHERE name = ?`, batchName).Scan(&bytes)
	if err != nil {
		return res, fmt.Errorf("failed to exec query : %v", err)
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't unmarshal batch properties : %v", err)
	}
	return res, nil
}

func GetBatchProperties(batchName string) (protocol.BatchMetadata, error) {
	var res protocol.BatchMetadata

	tx, err := Begin()
	if err != nil {
		return res, fmt.Errorf("failed to start transaction : %v", err)
	}

	res, err = GetBatchPropertiesTx(tx, batchName)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("no batch properties for %s: %v", batchName, err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

func SetBatchPropertiesTx(tx *sql.Tx, batchName string, properties []byte) error {
	_, err := tx.Exec("INSERT INTO batch_properties (name, properties) VALUES (?, ?)", batchName, properties)
	if err != nil {
		return fmt.Errorf("failed to exec query : %v", err)
	}
	return nil
}

func SetBatchProperties(batchName string, properties []byte) error {
	tx, err := Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction : %v", err)
	}

	err = SetBatchPropertiesTx(tx, batchName, properties)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("couldn't set batch properties: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return nil
}

// TODO: replaces question marks inside values as well, e.g. NOT REGEXP '[^a-zA-Z?]'
func PopulateQueryString(query string, args []interface{}) string {
	pop := query
	for _, arg := range args {
		switch arg.(type) {
		case string:
			pop = strings.Replace(pop, "?", fmt.Sprintf("'%v'", arg), 1)
			//pop = popQRE.ReplaceAllString(pop, fmt.Sprintf("${1}'%v${2}'", arg))
		default:
			pop = strings.Replace(pop, "?", fmt.Sprintf("%v", arg), 1)
			//pop = popQRE.ReplaceAllString(pop, fmt.Sprintf("${1}%v${2}", arg))

		}
	}
	return pop
}

// GetScriptTx retrieves all sentences in the specified script, using pageNumber and pageSize if provided.
// If pageNumber is zero and pageSize is zero, no limit settings are used.
// If pageNumber is set to zero, and pageSize is non-zero, the pageSize will be used as LIMIT in the query.
func GetScriptTx(tx *sql.Tx, scriptName string, pageNumber, pageSize int) ([]text.Sentence, error) {
	var res []text.Sentence
	var rows *sql.Rows
	var err error

	if pageNumber == 0 {
		if pageSize == 0 {
			rows, err = tx.Query("SELECT chunk_id FROM script WHERE script.name = ?", scriptName)
		} else {
			rows, err = tx.Query("SELECT chunk_id FROM script WHERE script.name = ? LIMIT ?", scriptName, pageSize)
		}
	} else {
		offset := (pageNumber - 1) * pageSize
		rows, err = tx.Query("SELECT chunk_id FROM script WHERE script.name = ? LIMIT ? OFFSET ?", scriptName, pageSize, offset)
	}

	if err != nil {
		return res, fmt.Errorf("failed to exec query : %v", err)
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	res, err = GetSentsTx(tx, ids...)
	if err != nil {
		return res, fmt.Errorf("failed to get sents from ids : %v", err)
	}
	return res, nil
}

// GetScript retrieves all sentences in the specified script, using pageNumber and pageSize if provided.
// If pageNumber is zero and pageSize is zero, no limit settings are used.
// If pageNumber is set to zero, and pageSize is non-zero, the pageSize will be used as LIMIT in the query.
func GetScript(scriptName string, pageNo, pageSize int) ([]text.Sentence, error) {
	var res []text.Sentence
	tx, err := Begin()
	if err != nil {
		return res, fmt.Errorf("failed to start transaction : %v", err)
	}

	res, err = GetScriptTx(tx, scriptName, pageNo, pageSize)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't get script: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

// GetBatchTx retrieves all sentences in the specified batch, using pageNumber and pageSize if provided.
// If pageNumber is zero and pageSize is zero, no limit settings are used.
// If pageNumber is set to zero, and pageSize is non-zero, the pageSize will be used as LIMIT in the query.
func GetBatchTx(tx *sql.Tx, batchName string, pageNumber, pageSize int) ([]text.Sentence, error) {
	var res []text.Sentence
	var rows *sql.Rows
	var err error

	if pageNumber == 0 {
		if pageSize == 0 {
			rows, err = tx.Query("SELECT chunk_id FROM batch WHERE batch.name = ?", batchName)
		} else {
			rows, err = tx.Query("SELECT chunk_id FROM batch WHERE batch.name = ? LIMIT ?", batchName, pageSize)
		}
	} else {
		offset := (pageNumber - 1) * pageSize
		rows, err = tx.Query("SELECT chunk_id FROM batch WHERE batch.name = ? LIMIT ? OFFSET ?", batchName, pageSize, offset)
	}

	if err != nil {
		return res, fmt.Errorf("failed to exec query : %v", err)
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	res, err = GetSentsTx(tx, ids...)
	if err != nil {
		return res, fmt.Errorf("failed to get sents from ids : %v", err)
	}
	return res, nil
}

// GetBatch retrieves all sentences in the specified batch, using pageNumber and pageSize if provided.
// If pageNumber is zero and pageSize is zero, no limit settings are used.
// If pageNumber is set to zero, and pageSize is non-zero, the pageSize will be used as LIMIT in the query.
func GetBatch(batchName string, pageNo, pageSize int) ([]text.Sentence, error) {
	var res []text.Sentence
	tx, err := Begin()
	if err != nil {
		return res, fmt.Errorf("failed to start transaction : %v", err)
	}

	res, err = GetBatchTx(tx, batchName, pageNo, pageSize)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't get batch: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

func SaveScript(metadata protocol.ScriptMetadata, sentIDs ...int64) (int, error) {
	tx, err := Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin db transaction : %v", err)
	}

	n := 0
	insertIntoScript := "INSERT INTO script (chunk_id, name) VALUES (?, ?)"
	for _, id := range sentIDs {
		_, err := tx.Exec(insertIntoScript, id, metadata.Options.ScriptName)
		if err != nil {
			tx.Rollback()
			return n, fmt.Errorf("failed to insert into script table : %v", err)
		}
		n++
	}

	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM script WHERE name = ?", metadata.Options.ScriptName).Scan(&count)
	if err != nil {
		tx.Rollback()
		return n, fmt.Errorf("failed to post-count script table : %v", err)
	}
	if count != n {
		tx.Rollback()
		return n, fmt.Errorf("SaveScript found %d sents, but output script in db contains %d", n, count)
	}

	metadata.OutputSize = count
	pBytes, err := json.Marshal(metadata)
	if err != nil {
		tx.Rollback()
		return n, fmt.Errorf("failed to marshal selection settings : %v", err)
	}
	err = SetScriptPropertiesTx(tx, metadata.Options.ScriptName, pBytes)
	if err != nil {
		tx.Rollback()
		return n, fmt.Errorf("failed to save script properties : %v", err)
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return n, fmt.Errorf("couldn't commit transaction: %v", err)
	}

	return n, nil
}
