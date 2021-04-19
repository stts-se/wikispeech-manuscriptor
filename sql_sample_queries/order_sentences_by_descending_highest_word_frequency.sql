PRAGMA foreign_keys = ON;

--- Orders sentences (chunks) according to the least frequent word (chunkfeat) in the sentence (the word with the highest wordfreq.id)


SELECT MIN(wordfreq.freq), chunk.text, source.name FROM chunk JOIN chunk_chunkfeat on chunk_chunkfeat.chunk_id = chunk.id JOIN wordfreq ON wordfreq.chunkfeat_id = chunk_chunkfeat.chunkfeat_id JOIN source_chunk on source_chunk.chunk_id = chunk.id JOIN source ON source_chunk.source_id = source.id GROUP BY chunk.id ORDER BY wordfreq.freq DESC LIMIT 5000;


-- select c.id from chunk as c join chunk_chunkfeat as ccf on c.id = ccf.chunk_id  join chunkfeat as cf on cf.id = ccf.chunkfeat_id and cf.name = 'digit'

-- With a min freq threshold


-- SELECT MIN(wordfreq.freq), chunk.text, source.name FROM chunk JOIN chunk_chunkfeat on chunk_chunkfeat.chunk_id = chunk.id JOIN wordfreq ON wordfreq.chunkfeat_id = chunk_chunkfeat.chunkfeat_id JOIN source_chunk on source_chunk.chunk_id = chunk.id JOIN source ON source_chunk.source_id = source.id GROUP BY chunk.id HAVING wordfreq.freq > 900 ORDER BY wordfreq.freq DESC LIMIT 20;


-- Example from http://joshualande.com/filters-joins-aggregations
--  SELECT b.recipe_name, 
--          COUNT(a.ingredient_id) AS num_ingredients
--     FROM recipe_ingredients AS a
--     JOIN recipes AS b
--       ON a.recipe_id = b.recipe_id
--     JOIN (
--              SELECT c.recipe_id
--              FROM recipe_ingredients AS c
--              JOIN ingredients AS d
--              ON c.ingredient_id = d.ingredient_id
--              WHERE d.ingredient_name = 'Tomatoes' 
--          ) AS e
--       ON b.recipe_id = e.recipe_id
-- GROUP BY a.recipe_id
