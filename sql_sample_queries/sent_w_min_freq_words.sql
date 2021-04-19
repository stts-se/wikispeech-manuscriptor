--- SLOW! Needs improvement
--- Broken without distinct

select distinct text from chunk as c join chunk_chunkfeat on c.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.name = 'word' and c.id not in (select chunk.id from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id  join chunkfeat, wordfreq on chunkfeat.id = chunk_chunkfeat.chunkfeat_id and chunkfeat.id = wordfreq.chunkfeat_id and wordfreq.id > 1000);

