-- sentence + source from articles of 8 paragraphs

select chunk.text, source.name from source, sourcefeat, source_sourcefeat, chunk, source_chunk where chunk.id = source_chunk.chunk_id and source.id = source_chunk.source_id and source.id = source_sourcefeat.source_id and sourcefeat.id = source_sourcefeat.sourcefeat_id and sourcefeat.name = "count" and sourcefeat.value = 'paragraph_count' and source_sourcefeat.freq = 8;

