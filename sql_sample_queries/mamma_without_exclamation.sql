-- NB: Slow!
-- inner select: all sentences including "mamma" but excluding "!"

--select text from chunk as c join chunk_chunkfeat on c.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value = 'mamma' and c.id not in (select chunk.id from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value like '%!%');

select text from chunk as c join chunk_chunkfeat on c.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value = 'mamma' and chunkfeat.name = 'word' and c.id not in (select chunk.id from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value = '!' and chunkfeat.name = 'punct');


--select text from chunk as c join chunk_chunkfeat on c.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value = 'mamma' and c.id not in (select chunk.id from chunk join chunk_chunkfeat on chunk.id = chunk_chunkfeat.chunk_id  join chunkfeat on chunkfeat.id = chunk_chunkfeat.chunkfeat_id where chunkfeat.value like '%!%');

--- https://stackoverflow.com/questions/4968570/how-to-select-the-relative-complement-of-a-table-b-in-a-table-a-a-b-in-an-sq
