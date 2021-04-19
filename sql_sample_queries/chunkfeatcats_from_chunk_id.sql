SELECT chunkfeatcat.name, chunkfeat.value FROM chunk, chunkfeat, chunkfeatcat, chunk_chunkfeat WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunkfeat.id = chunk_chunkfeat.chunkfeat_id AND chunkfeatcat.chunkfeat_id = chunkfeat.id AND chunk.id = 12;


-- SELECT chunk.id, chunk.text, chunkfeat.name, chunkfeat.value, chunk_chunkfeat.freq, chunkfeatcat.name FROM chunk, chunk_chunkfeat, chunkfeat LEFT JOIN chunkfeatcat ON chunkfeatcat.chunkfeat_id = chunkfeat.id WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunk_chunkfeat.chunkfeat_id = chunkfeat.id AND chunk.id = 12;
