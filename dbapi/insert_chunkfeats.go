package dbapi

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/stts-se/manuscriptor2000/text"
)

// TODO mutex this map?
var chunkFeatCache = map[string]map[string]int64{}

func InsertChunkFeatsTx(tx *sql.Tx, chunkID int64, feats map[string]map[string]int) error {

	var err error

	for fName, fVals := range feats {

		for fVal, freq := range fVals {

			// Check if feat+val is cached
			chunkFeatID := chunkFeatCache[fName][fVal]
			if chunkFeatID == 0 {
				var id sql.NullInt64
				err = tx.QueryRow("SELECT id FROM chunkfeat WHERE name = ? AND value = ?", fName, fVal).Scan(&id)
				chunkFeatID = id.Int64

				switch {
				case err == sql.ErrNoRows:
					execRes, err := tx.Exec("INSERT INTO chunkfeat(name, value)  VALUES(?, ?)", fName, fVal)
					if err != nil {
						//tx.Rollback()
						return fmt.Errorf("failed to insert chunkfeat %s %s", fName, fVal)
					}
					chunkFeatID, err = execRes.LastInsertId()
					if err != nil {
						//tx.Rollback()
						return fmt.Errorf("failed LastInserId() : %v", err)
					}

				case err != nil:
					//tx.Rollback()
					return fmt.Errorf("failed QueryRow : %v", err)

				}
				// Add new chunkfeat to cache
				if _, ok := chunkFeatCache[fName]; !ok {
					chunkFeatCache[fName] = map[string]int64{}
				}
				chunkFeatCache[fName][fVal] = chunkFeatID
			}
			// Now the chunkfeat is in the db, lets add the chunk - chunkfeat relation

			_, err := tx.Exec("INSERT OR IGNORE INTO chunk_chunkfeat(chunk_id, chunkfeat_id, freq) VALUES(?, ?, ?)", chunkID, chunkFeatID, freq)
			if err != nil {
				//tx.Rollback()
				return fmt.Errorf("failed to insert chunk_chunkfeat relation : %v", err)
			}
		}
	}

	return nil
}

const sqlite3Cmd = "sqlite3"

func sqlite3CmdExists() error {
	_, pErr := exec.LookPath(sqlite3Cmd)
	if pErr != nil {
		return fmt.Errorf("external '%s' command does not exist", sqlite3Cmd)
	}
	return nil
}

const debugBulkInsertChunkFeats = false

func BulkInsertChunkFeats(dbFile string, sents ...text.Sentence) (int, error) {
	n := 0

	tx, err := Begin()
	if err != nil {
		return n, fmt.Errorf("failed to begin db transaction : %v", err)
	}

	err = ForeignKeysOff()
	if err != nil {
		return n, fmt.Errorf("failed ForeignKeysOff : %v", err)
	}

	err = SynchronousOff()
	if err != nil {
		return n, fmt.Errorf("failed SynchronousOff : %v", err)
	}

	err = CacheSize(10000)
	if err != nil {
		return n, fmt.Errorf("failed CacheSize : %v", err)
	}

	defer func() {
		err = ForeignKeysOn()
		if err != nil {
			log.Printf("Failed ForeignKeysOn : %v", err)
		}

		err = SynchronousOn()
		if err != nil {
			log.Printf("Failed SynchronousOff : %v", err)
		}
	}()

	// CREATE TMP BUFFER
	tmpFN := filepath.Join(fmt.Sprintf("dbapi-bulkinsertchunkfeats-%s.tsv", text.RandomString(10)))
	tmpFile, err := os.Create(tmpFN)
	if err != nil {
		return n, fmt.Errorf("failed to create temp file : %v", err)
	}
	if debugBulkInsertChunkFeats {
		log.Printf("[debug] Writing tab sep to %s", tmpFile.Name())
	}
	tmpFW := bufio.NewWriter(tmpFile)
	if err != nil {
		return n, fmt.Errorf("failed to open file writer %s : %v", tmpFile.Name(), err)
	}
	defer func() {
		tmpFW.Flush()
		tmpFile.Close()
		if !debugBulkInsertChunkFeats {
			os.Remove(tmpFile.Name())
		}
	}()

	// LOOP
	for _, sent := range sents {
		for fName, fVals := range sent.Feats {

			for fVal, freq := range fVals {
				n++

				// Check if feat+val is cached
				chunkFeatID := chunkFeatCache[fName][fVal]
				if chunkFeatID == 0 {
					var id sql.NullInt64
					err = tx.QueryRow("SELECT id FROM chunkfeat WHERE name = ? AND value = ?", fName, fVal).Scan(&id)
					chunkFeatID = id.Int64

					switch {
					case err == sql.ErrNoRows:
						execRes, err := tx.Exec("INSERT INTO chunkfeat(name, value) VALUES(?, ?)", fName, fVal)
						if err != nil {
							//tx.Rollback()
							return n, fmt.Errorf("failed to insert chunkfeat %s %s", fName, fVal)
						}
						chunkFeatID, err = execRes.LastInsertId()
						if err != nil {
							//tx.Rollback()
							return n, fmt.Errorf("failed LastInseryId() : %v", err)
						}

					case err != nil:
						//tx.Rollback()
						return n, fmt.Errorf("failed QueryRow : %v", err)

					}
					// Add new chunkfeat to cache
					if _, ok := chunkFeatCache[fName]; !ok {
						chunkFeatCache[fName] = map[string]int64{}
					}
					chunkFeatCache[fName][fVal] = chunkFeatID
				}

				// Now the chunkfeat is in the db, lets add the chunk - chunkfeat relation

				// https://stackoverflow.com/questions/364017/faster-bulk-inserts-in-sqlite3
				tmpFW.Write([]byte(fmt.Sprintf("%v\t%v\t%v\n", sent.ID, chunkFeatID, freq)))
			}
		}
	}

	// IMPORT FROM TMP BUFFER

	err = tx.Commit()
	if err != nil {
		return n, fmt.Errorf("couldn't commit transaction : %v", err)
	}

	tmpFW.Flush()
	tmpFile.Close()

	cmdLines := []string{
		`PRAGMA foreign_keys =OFF;`,
		`PRAGMA synchronous = OFF;`,
		`PRAGMA cache_size=10000;`,
		`.mode csv`,
		`.separator '	'`,
		`.headers off`,
		fmt.Sprintf(`.import %s chunk_chunkfeat`, tmpFile.Name()),
	}
	if sqlite3CmdExists(); err != nil {
		return 0, err
	}
	subProcess := exec.Command(sqlite3Cmd, dbFile)

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		return n, fmt.Errorf("failed open sqlite3 subprocess: %v", err)
	}
	defer stdin.Close()

	var stderr bytes.Buffer
	subProcess.Stdout = &stderr
	subProcess.Stderr = &stderr

	//fmt.Fprint(os.Stderr, "Importing bulk... ")
	if err = subProcess.Start(); err != nil {
		return n, fmt.Errorf("failed to start sqlite3 subprocess : %v", err)
	}
	for _, l := range cmdLines {
		io.WriteString(stdin, fmt.Sprintf("%s\n", l))
	}
	stdin.Close()
	err = subProcess.Wait()
	if err != nil {
		return n, fmt.Errorf("failed to run sqlite3 subprocess : %v (%v)", strings.TrimSpace(string(stderr.Bytes())), err)
	}
	//fmt.Fprintln(os.Stderr, "done")
	return n, nil
}

func bulkInsertChunkFeatsTx(tx *sql.Tx, chunkID int64, feats map[string]map[string]int, tmpFW *bufio.Writer) error {

	var err error

	for fName, fVals := range feats {

		for fVal, freq := range fVals {

			// Check if feat+val is cached
			chunkFeatID := chunkFeatCache[fName][fVal]
			if chunkFeatID == 0 {
				var id sql.NullInt64
				err = tx.QueryRow("SELECT id FROM chunkfeat WHERE name = ? AND value = ?", fName, fVal).Scan(&id)
				chunkFeatID = id.Int64

				switch {
				case err == sql.ErrNoRows:
					execRes, err := tx.Exec("INSERT INTO chunkfeat(name, value)  VALUES(?, ?)", fName, fVal)
					if err != nil {
						//tx.Rollback()
						return fmt.Errorf("failed to insert chunkfeat %s %s", fName, fVal)
					}
					chunkFeatID, err = execRes.LastInsertId()
					if err != nil {
						//tx.Rollback()
						return fmt.Errorf("failed LastInserId() : %v", err)
					}

				case err != nil:
					//tx.Rollback()
					return fmt.Errorf("failed QueryRow : %v", err)

				}
				// Add new chunkfeat to cache
				if _, ok := chunkFeatCache[fName]; !ok {
					chunkFeatCache[fName] = map[string]int64{}
				}
				chunkFeatCache[fName][fVal] = chunkFeatID
			}
			// Now the chunkfeat is in the db, lets add the chunk - chunkfeat relation

			// https://stackoverflow.com/questions/364017/faster-bulk-inserts-in-sqlite3
			tmpFW.Write([]byte(fmt.Sprintf("%v\t%v\t%v\n", chunkID, chunkFeatID, freq)))
		}
	}

	return nil
}
