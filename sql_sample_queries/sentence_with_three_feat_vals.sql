-- NB: Slow!
--- Select all sentences with three specific feature values
--- Cf https://stackoverflow.com/questions/10888097/sql-query-for-join-table-and-multiple-values
--- https://stackoverflow.com/questions/36131803/sql-where-joined-set-must-contain-all-values-but-may-contain-more 


-- select chunk.text from chunk, chunkfeat, chunk_chunkfeat where chunkfeat.value in ('mamma', 'pappa', 'hund') and chunkfeat.id = chunk_chunkfeat.chunkfeat_id AND chunk.id = chunk_chunkfeat.chunk_id group by chunk.text having count(chunkfeat.value) = 3;

-- Variant:
select chunk.text from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id and chunkfeat.name = 'word' and chunkfeat.value in ('mamma', 'pappa', 'hund') GROUP BY chunk.id HAVING COUNT(chunkfeat.value) = 3;

--- At least 3 of 8 words
--- select chunk.text from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id and chunkfeat.name = 'word' and chunkfeat.value in ('mamma', 'pappa', 'hund', 'katt', 'skog', 'smitta', 'miljö', 'häst') GROUP BY chunk.id HAVING COUNT(chunkfeat.value) > 2;

-- "Echinococcus multilocularis" sprids genom så kallad sylvatisk smitta, dvs. mellan vilda djur, men kan även spridas via tamdjur. 
