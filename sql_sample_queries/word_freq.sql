--- Generating fresh list without wordfreq table
--- NB: This takes _time_ on larger DB
--select chunk_chunkfeat.chunkfeat_id, chunkfeat.value, count(*) from chunk_chunkfeat, chunkfeat where chunkfeat.name = "word" and chunkfeat.id = chunk_chunkfeat.chunkfeat_id group by chunk_chunkfeat.chunkfeat_id order by count(chunk_chunkfeat.chunk_id) DESC limit 25;


-- TOPLIST
select chunkfeat.value, wordfreq.id, wordfreq.freq from wordfreq, chunkfeat where wordfreq.chunkfeat_id = chunkfeat.id ORDER BY wordfreq.id LIMIT 40;

-- Specific word
-- select chunkfeat.value, wordfreq.id, wordfreq.freq from wordfreq, chunkfeat where wordfreq.chunkfeat_id = chunkfeat.id and chunkfeat.value = 'även' ORDER BY wordfreq.id;

--- Selecting top of frequency list. wordfreq is ordered on insert. 
-- select chunkfeat.value, wordfreq.freq from wordfreq, chunkfeat where wordfreq.chunkfreq_id = chunkfeat.id limit 25;

