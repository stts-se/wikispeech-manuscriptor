package dbapi

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/stts-se/manuscriptor2000/text"
)

type ChunkFeatCat struct {
	FeatValue      string
	TargetFeatName string
}

func ParseChunkFeatCatFile(fn string) (string, []ChunkFeatCat, error) {
	var res []ChunkFeatCat
	var sourceFeatName string

	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return sourceFeatName, res, fmt.Errorf("func lines() failed to read file '%s' : %v", fn, err)
	}

	lines := strings.Split(strings.TrimSuffix(string(bytes), "\n"), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "#") {
			log.Printf("skipping commented line: '%s'", l)
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		fs := strings.Split(l, "\t")
		if len(fs) != 3 {
			log.Printf("skipping faulty line: '%s'", l)
			continue
		}
		fn := strings.ToLower(strings.TrimSpace(fs[0]))

		if sourceFeatName != "" && fn != sourceFeatName {
			return sourceFeatName, res, fmt.Errorf("mixed source feat names: %s, %s in file %s", sourceFeatName, fn, fn)
		}
		sourceFeatName = fn

		cat := strings.ToLower(strings.TrimSpace(fs[1]))
		value := strings.ToLower(strings.TrimSpace(fs[2]))

		if cat == "" || value == "" {
			log.Printf("skipping faulty line: '%s'", l)
			continue

		}

		res = append(res, ChunkFeatCat{TargetFeatName: cat, FeatValue: value})
	}

	return sourceFeatName, res, nil
}

// AddChunkFeatCats takes a list of ChunkFeatCats and inserts the associted FEAT in the chunkfeatcat relation table.
func AddChunkFeatCats(sourceFeatName string, feats []ChunkFeatCat) (int, error) {

	tmpTableName := fmt.Sprintf("chunk_feats_to_add_%s", text.RandomString(10))

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("AddChunkFeatFeats failed to start transaction : %v", err)
	}

	_, err = tx.Exec(`CREATE TEMP TABLE IF NOT EXISTS ` + tmpTableName + ` (name TEXT NOT NULL, value TEXT NOT NULL, UNIQUE(name, value))`)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create temp db table : %v", err)
	}

	featNameMap := make(map[string]map[string]bool)
	for _, f := range feats {
		name := strings.ToLower(strings.TrimSpace(f.TargetFeatName))
		value := strings.ToLower(strings.TrimSpace(f.FeatValue))
		if name == "" || value == "" {
			log.Printf("empty field, skipping feature %v : ", f)
			continue
		}

		if _, ok := featNameMap[value]; !ok {
			featNameMap[value] = map[string]bool{}
		}
		featNameMap[value][name] = true

		_, err := tx.Exec(`INSERT OR IGNORE INTO `+tmpTableName+`(name, value) VALUES (?, ?)`, name, value)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to insert feature %s-%s into temp table : %v", f.TargetFeatName, f.FeatValue, err)
		}

	}

	rows, err := tx.Query(`SELECT chunkfeat.id, chunkfeat.value FROM chunkfeat WHERE chunkfeat.name = ? AND chunkfeat.value IN (SELECT value FROM `+tmpTableName+`)`, sourceFeatName)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed db query : %v", err)
	}

	n := 0
	for rows.Next() {
		var chunkfeatID sql.NullInt64
		var val string

		err := rows.Scan(&chunkfeatID, &val)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to scan row : %v", err)
		}

		q := `INSERT OR IGNORE INTO chunkfeatcat (name, chunkfeat_id) VALUES (?, ?)`

		for name := range featNameMap[val] {

			xRes, err := tx.Exec(q, name, chunkfeatID.Int64)
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("failed to insert into chunkfeatcat table : %v", err)
			}
			ra, err := xRes.RowsAffected()
			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("failed call to RowsAffected : %v", err)
			}

			if ra > 0 {
				n++
			}
		}

	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("couldn't commit transaction : %v", err)
	}

	return n, nil
}
