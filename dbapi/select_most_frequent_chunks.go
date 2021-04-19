package dbapi

import (
	"fmt"
)

type Chunk struct {
	Freq int64
	Text string
	ID   int64
}

var chunkFreqQuery = `SELECT chunk.id, chunk.text, COUNT(source.id) sfreq FROM source, chunk, source_chunk WHERE chunk.id = source_chunk.chunk_id AND source.id = source_chunk.source_id GROUP BY chunk.id ORDER BY sfreq DESC LIMIT ?`

// SelectMostFrequentChunks lists the chunks in the DB sorted
// according to in how many sources (articles) they occurr in. It is
// useful mostly for diagnostic purposes.
func SelectMostFrequentChunks(limit int) ([]Chunk, error) {
	var res []Chunk

	rows, err := db.Query(chunkFreqQuery, limit)
	if err != nil {
		return res, fmt.Errorf("SelectMostFrequentChunks failed query DB : %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, freq int64
		var chunk string
		err = rows.Scan(&id, &chunk, &freq)
		if err != nil {
			return res, fmt.Errorf("SelectMostFrequentChunks failed scanning row : %v", err)
		}

		res = append(res, Chunk{ID: id, Text: chunk, Freq: freq})
	}

	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("SelectMostFrequentChunks failed : %v", err)
	}

	return res, nil
}
