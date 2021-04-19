package filter

// inspired by https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
// and https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html

import (
	"fmt"
	"strings"

	"github.com/stts-se/wikispeech-manuscriptor/dbapi"
	"github.com/stts-se/wikispeech-manuscriptor/text"
)

type queryBuilder struct {
	head  string
	joins []string
	tail  string
	args  []interface{}
}

func (qb *queryBuilder) query() (string, []interface{}) {

	qString := qb.head + " " + strings.Join(qb.joins, " ") + " " + qb.tail

	return qString, qb.args
}

func (qb *queryBuilder) populatedQueryString() string {
	query, args := qb.query()
	return dbapi.PopulateQueryString(query, args)
}

type opt func(*queryBuilder)

func newFilterQueryBuilder(options ...opt) (*queryBuilder, error) {
	res := &queryBuilder{}
	for _, o := range options {
		o(res)
	}
	// todo: validering
	return res, nil
}

func filterQHeadInto(intoBatch string) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		qb.head = `INSERT OR IGNORE INTO batch (chunk_id, name) SELECT DISTINCT chunk.id, ? FROM chunk`
		qb.args = append(qb.args, intoBatch)
	}
}

func fromBatchQ(batchName string) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN batch ON chunk.id = batch.chunk_id AND batch.name = ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, batchName)
	}
}

func fromSourceRE(sourceRE string) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN source, source_chunk ON chunk.id = source_chunk.chunk_id AND source.id = source_chunk.source_id AND source.name REGEXP ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, sourceRE)
	}
}

func innerJoinSourcefeatCount(countType string, operand string, count int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		rid := text.RandomString(10)
		sourceTbl := fmt.Sprintf("source_%s", rid)
		sourceChunkTbl := fmt.Sprintf("source_chunk_%s", rid)
		sourcefeatTbl := fmt.Sprintf("sourcefeat_%s", rid)
		sourceSourcefeatTbl := fmt.Sprintf("source_sourcefeat_%s", rid)
		j := fmt.Sprintf(`JOIN source AS %s, source_chunk AS %s, sourcefeat AS %s, source_sourcefeat AS %s ON chunk.id = %s.chunk_id AND %s.id = %s.source_id AND %s.id = %s.sourcefeat_id AND %s.id = %s.source_id AND %s.value = ? AND %s.name = 'count' AND %s.freq %s ?`,
			sourceTbl, sourceChunkTbl, sourcefeatTbl, sourceSourcefeatTbl,
			sourceChunkTbl,
			sourceTbl,
			sourceChunkTbl,
			sourcefeatTbl,
			sourceSourcefeatTbl,
			sourceTbl,
			sourceSourcefeatTbl,
			sourcefeatTbl,
			sourcefeatTbl,
			sourceSourcefeatTbl,
			operand,
		)
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, countType)
		qb.args = append(qb.args, count)
	}
}

func innerJoinSourcefeatCountInterval(countType string, min, max int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		rid := text.RandomString(10)
		sourceTbl := fmt.Sprintf("source_%s", rid)
		sourceChunkTbl := fmt.Sprintf("source_chunk_%s", rid)
		sourcefeatTbl := fmt.Sprintf("sourcefeat_%s", rid)
		sourceSourcefeatTbl := fmt.Sprintf("source_sourcefeat_%s", rid)
		j := fmt.Sprintf(`JOIN source AS %s, source_chunk AS %s, sourcefeat AS %s, source_sourcefeat AS %s ON chunk.id = %s.chunk_id AND %s.id = %s.source_id AND %s.id = %s.sourcefeat_id AND %s.id = %s.source_id AND %s.value = ? AND %s.name = 'count' AND %s.freq >= ? AND %s.freq <= ?`,
			sourceTbl, sourceChunkTbl, sourcefeatTbl, sourceSourcefeatTbl,
			sourceChunkTbl,
			sourceTbl,
			sourceChunkTbl,
			sourcefeatTbl,
			sourceSourcefeatTbl,
			sourceTbl,
			sourceSourcefeatTbl,
			sourcefeatTbl,
			sourcefeatTbl,
			sourceSourcefeatTbl,
			sourceSourcefeatTbl,
		)
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, countType)
		qb.args = append(qb.args, min)
		qb.args = append(qb.args, max)
	}
}

func paragraphCountInterval(min, max int) func(*queryBuilder) {
	return innerJoinSourcefeatCountInterval("paragraph_count", min, max)
}

func sentenceCountInterval(min, max int) func(*queryBuilder) {
	return innerJoinSourcefeatCountInterval("sentence_count", min, max)
}

func paragraphCountMin(count int) func(*queryBuilder) {
	return innerJoinSourcefeatCount("paragraph_count", ">=", count)
}

func paragraphCountMax(count int) func(*queryBuilder) {
	return innerJoinSourcefeatCount("paragraph_count", "<=", count)
}

func sentenceCountMin(count int) func(*queryBuilder) {
	return innerJoinSourcefeatCount("sentence_count", ">=", count)
}

func sentenceCountMax(count int) func(*queryBuilder) {
	return innerJoinSourcefeatCount("sentence_count", "<=", count)
}

func fromScriptQ(scriptName string) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN script ON chunk.id = script.chunk_id AND script.name = ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, scriptName)
	}
}

func wordCountView(min, max int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN chunk_wordcount_view ON chunk.id = chunk_wordcount_view.id AND chunk_wordcount_view.freq >= ? AND chunk_wordcount_view.freq <= ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, min)
		qb.args = append(qb.args, max)
	}
}

func commaCountView(min, max int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN chunk_commacount_view ON chunk.id = chunk_commacount_view.id AND chunk_commacount_view.freq >= ? AND chunk_commacount_view.freq <= ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, min)
		qb.args = append(qb.args, max)
	}
}

// experimental! only works with SELECT DISTINCT, and barely even then
// func excludePuncts(puncts ...string) func(*queryBuilder) {
// 	return func(qb *queryBuilder) {
// 		var qs []string
// 		for i := 0; i < len(puncts); i++ {
// 			qs = append(qs, "?")
// 		}
// 		j := "JOIN chunk_chunkfeat AS ccf_ep, chunkfeat AS cf_ep ON chunk.id = ccf_ep.chunk_id AND cf_ep.id = ccf_ep.chunkfeat_id AND chunk.id NOT IN (SELECT c_ep0.id FROM chunk AS c_ep0 JOIN chunk_chunkfeat AS ccf_ep0, chunkfeat AS cf_ep0 ON c_ep0.id = ccf_ep0.chunk_id AND cf_ep0.id = ccf_ep0.chunkfeat_id AND cf_ep0.value IN ( " + strings.Join(qs, ", ") + " ) AND cf_ep0.name = 'punct')"
// 		qb.joins = append(qb.joins, j)
// 		for _, p := range puncts {
// 			qb.args = append(qb.args, p)
// 		}
// 	}
// }

func excludeChunkLike(re string) func(*queryBuilder) {
	tableName := fmt.Sprintf("chunk_ecl_%s", text.RandomString(10))
	return func(qb *queryBuilder) {
		j := fmt.Sprintf(`JOIN chunk AS %s ON chunk.id = %s.id AND %s.text NOT LIKE ?`, tableName, tableName, tableName)
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, re)
	}
}

func excludeChunkRE(re string) func(*queryBuilder) {
	tableName := fmt.Sprintf("chunk_ecre_%s", text.RandomString(10))
	return func(qb *queryBuilder) {
		j := fmt.Sprintf(`JOIN chunk AS %s ON chunk.id = %s.id AND %s.text NOT REGEXP ?`, tableName, tableName, tableName)
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, re)
	}
}

func nDigitCountView(n int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN chunk_digitcount_view ON chunk.id = chunk_digitcount_view.id AND chunk_digitcount_view.freq = ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, n)
	}
}

func digitCountView(gt, lt int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN chunk_digitcount_view ON chunk.id = chunk_digitcount_view.id AND chunk_digitcount_view.freq >= ? AND chunk_digitcount_view.freq <= ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, gt)
		qb.args = append(qb.args, lt)
	}
}

func lowestWordFreqCountView(lowFreq int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := `JOIN chunk_lowestwordfreqcount_view ON chunk.id = chunk_lowestwordfreqcount_view.id AND chunk_lowestwordfreqcount_view.freq > ?`
		qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, lowFreq)
	}
}

func chunkFeatCat(featCatNames ...string) func(*queryBuilder) {
	return func(qb *queryBuilder) {

		var qs []string
		for i := 0; i < len(featCatNames); i++ {
			qs = append(qs, "?")
		}

		// j := "JOIN chunk_chunkfeat AS ccf ON chunk.id = ccf.chunk_id JOIN chunkfeat AS cf ON cf.id = ccf.chunkfeat_id AND chunk.id NOT IN (SELECT c.id FROM chunk AS c JOIN chunk_chunkfeat AS ccfb, chunkfeat AS cfb ON c.id = ccfb.chunk_id AND cfb.id = ccfb.chunkfeat_id AND cfb.value IN ( " + strings.Join(qs, ", ") + " AND cfb.name = 'punct')"
		j := `JOIN chunkfeatcat, chunkfeat, chunk_chunkfeat ON chunk.id = chunk_chunkfeat.chunk_id and chunkfeat.id = chunk_chunkfeat.chunkfeat_id and chunkfeat.id = chunkfeatcat.chunkfeat_id and chunkfeatcat.name IN (` + strings.Join(qs, ", ") + `)`
		qb.joins = append(qb.joins, j)
		for _, fn := range featCatNames {
			qb.args = append(qb.args, fn)
		}
	}
}

func tailNotInBatches(batchNames ...string) func(*queryBuilder) {
	var qs = []string{"?"}
	for i := 0; i < len(batchNames); i++ {
		qs = append(qs, "?")
	}
	return func(qb *queryBuilder) {
		t := `WHERE chunk.id NOT IN ( SELECT chunk_id FROM batch WHERE name IN (` + strings.Join(qs, ", ") + `) )`
		qb.tail = t
		//qb.joins = append(qb.joins, j)
		qb.args = append(qb.args, text.BlockBatch)
		for _, bn := range batchNames {
			qb.args = append(qb.args, bn)
		}
	}
}

// func tailNotInBatchOrderByLowestWordFreq(notInBatch string, limit int) func(*queryBuilder) {
// 	return func(qb *queryBuilder) {
// 		j := `WHERE chunk.id NOT IN ( SELECT chunk_id FROM batch WHERE name = IN (?,?) ) ORDER BY chunk_lowestwordfreqcount_view.freq DESC LIMIT ?`
// 		qb.tail = j
// 		qb.args = append(qb.args, notInBatch)
// 		qb.args = append(qb.args, text.BlockBatch)
// 		qb.args = append(qb.args, limit)
// 	}
// }

func tailLimit(limit int) func(*queryBuilder) {
	return func(qb *queryBuilder) {
		j := ` LIMIT ?`
		qb.tail += j
		qb.args = append(qb.args, limit)
	}
}
