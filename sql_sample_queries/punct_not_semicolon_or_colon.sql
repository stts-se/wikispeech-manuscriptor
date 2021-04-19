-- work in progress

SELECT c.id, text from chunk as c join chunk_chunkfeat on c.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where c.id not in (select chunk.id from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value in ( ';', ':' ) and chunkfeat.name = 'punct');
