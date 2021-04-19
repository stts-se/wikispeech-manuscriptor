package filter

import (
	"fmt"
	"os"

	"github.com/stts-se/manuscriptor2000/dbapi"
)

const debug = true

func ExecQuery(qb *queryBuilder) (int64, error) {
	var res int64

	qString, args := qb.query()

	if debug {
		// log.Printf("filter debug qString\t%s\n", qString)
		// log.Printf("filter debug args\t%#v\n", args)
		fmt.Fprintf(os.Stderr, "[filter] Populated query %s\n", qb.populatedQueryString())
	}

	tx, err := dbapi.Begin()
	if err != nil {
		return res, fmt.Errorf("failed to begin transaction : %v", err)
	}

	result, err := tx.Exec(qString, args...)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed to execute query %v: %v", qString, err)
	}

	res, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("failed RowsAffected call to query Result : %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return res, fmt.Errorf("couldn't commit transaction: %v", err)
	}
	return res, nil
}

// func FilterFeatCatFromBatchIntoBatch(fromBatch, toBatch, excludeBatch, featName string, n int) (int64, error) {
// 	opts := []opt{
// 		filterQHeadInto(toBatch),
// 		fromBatchQ(fromBatch),
// 		chunkFeatCat(featName),
// 	}
// 	if n > 0 {
// 		opts = append(opts, tailLimit(n))
// 	}
// 	qb, err := newFilterQueryBuilder(opts...)
// 	if err != nil {
// 		return 0, fmt.Errorf("couldn't create query builder : %v", err)
// 	}
// 	return ExecQuery(qb)
// }

// func BasicFilterIntoBatch(toBatch, excludeBatch string, n int) (int64, error) {
// 	opts := []opt{
// 		filterQHeadInto(toBatch),
// 		wordCountView(4, 25),
// 		commaCountView(0, 3),
// 		nDigitCountView(0),
// 		lowestWordFreqCountView(2),
// 		//excludePuncts(";"),
// 		excludeChunkRE(`[\p{Greek}]`),
// 		excludeChunkRE(`[^a-zA-ZåäöÅÄÖéÉüÜ0-9 ,$€£@.!?/()"':—–-]`),
// 		excludeChunkLike("%;%"),
// 		tailNotInBatches(excludeBatch),
// 		//tailOrderByLowestWordFreq(),
// 		//tailNotInBatchOrderByLowestWordFreq(excludeBatch, n),
// 	}
// 	if n > 0 {
// 		opts = append(opts, tailLimit(n))
// 	}
// 	qb, err := newFilterQueryBuilder(opts...)

// 	if err != nil {
// 		return 0, fmt.Errorf("couldn't create query builder : %v", err)
// 	}

// 	return ExecQuery(qb)
// }
