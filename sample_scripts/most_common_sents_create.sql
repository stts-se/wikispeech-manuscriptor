INSERT INTO script (chunk_id, name) SELECT chunk_id, 'most_common_sents_script_1' FROM source_chunk GROUP BY chunk_id
HAVING COUNT(*)>1
ORDER BY COUNT(*) DESC
LIMIT 1000;



