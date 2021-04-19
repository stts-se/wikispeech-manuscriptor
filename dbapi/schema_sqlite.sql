
-- Source is the text (article, etc) from which a text chunk is from.
-- INSERT INTO source (name) VALUES('source name...');
CREATE TABLE IF NOT EXISTS source(
       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
       name TEXT UNIQUE NOT NULL       
       );

CREATE UNIQUE INDEX source_indx on source(name);	


-- Sourcefea: features of a source
CREATE TABLE IF NOT EXISTS sourcefeat (
       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,	
       name TEXT NOT NULL,
       value TEXT NOT NULL
       );

CREATE INDEX IF NOT EXISTS sourcefeat_indx ON sourcefeat(value, name);
CREATE INDEX IF NOT EXISTS sourcefeatname_indx ON sourcefeat(name);
CREATE INDEX IF NOT EXISTS sourcefeatvalue_indx ON sourcefeat(value);

-- Chunk is a minimal piece of text, typical a sentence 
CREATE TABLE IF NOT EXISTS chunk(
       id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
       text TEXT NOT NULL UNIQUE
       );

CREATE INDEX IF NOT EXISTS chnk_indx ON chunk(text);

-- Chunkfeat: features of a chunk (sentence)

CREATE TABLE IF NOT EXISTS chunkfeat (
       id INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,	
       name TEXT NOT NULL,
       value TEXT NOT NULL,
       UNIQUE(name, value)
       ); 

CREATE UNIQUE INDEX IF NOT EXISTS chnkfeat_indx ON chunkfeat(value, name);
CREATE INDEX IF NOT EXISTS chnkfeatname_indx ON chunkfeat(name);
CREATE INDEX IF NOT EXISTS chnkfeatvalue_indx ON chunkfeat(value);

CREATE TABLE IF NOT EXISTS batch_properties (
       id INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,	
       name TEXT NOT NULL,
       properties TEXT NOT NULL,
       UNIQUE(name)
       ); 

CREATE TABLE IF NOT EXISTS script_properties (
       id INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,	
       name TEXT NOT NULL,
       properties TEXT NOT NULL,
       UNIQUE(name)
       ); 



---
-- Relational linking tables
---

-- chunkfeatcat is a small feline.
-- Associates a category to a specific chunkfeat.
-- Example: chunkfeat name:word value:stockholm -> chunkfeatcat name:se_city 
CREATE TABLE IF NOT EXISTS chunkfeatcat(
       name TEXT NOT NULL,
       chunkfeat_id INTEGER NOT NULL,
       UNIQUE(name, chunkfeat_id),
       FOREIGN KEY (chunkfeat_id) REFERENCES chunkfeat(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS chunkfeatcat_name ON chunkfeatcat(name);
CREATE UNIQUE INDEX IF NOT EXISTS chunkfeatcat_name_cfid ON chunkfeatcat(name, chunkfeat_id);
CREATE INDEX IF NOT EXISTS chunkfeatcat_cfid ON chunkfeatcat(chunkfeat_id);

CREATE TABLE IF NOT EXISTS source_chunk(
	     source_id INTEGER NOT NULL,
	     chunk_id INTEGER NOT NULL,
	     UNIQUE(source_id, chunk_id),
	     FOREIGN KEY (source_id) REFERENCES source(id) ON DELETE CASCADE,
	     FOREIGN KEY (chunk_id) REFERENCES chunk(id) ON DELETE CASCADE
	     );

CREATE UNIQUE INDEX IF NOT EXISTS source_chunk_idx ON source_chunk(source_id, chunk_id);
CREATE INDEX IF NOT EXISTS source_chunk_idx_2 ON source_chunk(source_id);
CREATE INDEX IF NOT EXISTS source_chunk_idx_3 ON source_chunk(chunk_id);
       

CREATE TABLE IF NOT EXISTS source_sourcefeat(
       source_id INTEGER NOT NULL,
       sourcefeat_id INTEGER NOT NULL,
       freq INTEGER NOT NULL,
       FOREIGN KEY (source_id) REFERENCES source(id) ON DELETE CASCADE,
       FOREIGN KEY (sourcefeat_id) REFERENCES sourcefeat(id) ON DELETE CASCADE
       );

CREATE INDEX IF NOT EXISTS source_sourcefeat_indx ON source_sourcefeat(sourcefeat_id, source_id);
CREATE INDEX IF NOT EXISTS source_sourcefeat_indx_2 ON source_sourcefeat(sourcefeat_id);
CREATE INDEX IF NOT EXISTS source_sourcefeat_indx_3 ON source_sourcefeat(source_id);

CREATE TABLE IF NOT EXISTS chunk_chunkfeat(
       	     chunk_id INTEGER NOT NULL,
	     chunkfeat_id INTEGER NOT NULL,
	     freq INTEGER NOT NULL DEFAULT 0,
	     UNIQUE(chunkfeat_id, chunk_id, freq),
	     foreign key (chunk_id) references chunk(id) ON DELETE CASCADE,
	     foreign key (chunkfeat_id) references chunkfeat(id) ON DELETE CASCADE
       );

CREATE UNIQUE INDEX IF NOT EXISTS chunk_chunkfeat_freq_indx ON chunk_chunkfeat(chunkfeat_id, chunk_id, freq);
CREATE INDEX IF NOT EXISTS chunk_chunkfeat_indx ON chunk_chunkfeat(chunkfeat_id, chunk_id);
CREATE INDEX IF NOT EXISTS chunk_chunkfeat_chunk_id_indx ON chunk_chunkfeat(chunk_id);
CREATE INDEX IF NOT EXISTS chunk_chunkfeat_chunkfeat_id_indx ON chunk_chunkfeat(chunkfeat_id);
CREATE INDEX IF NOT EXISTS chunk_chunkfeat_freq_id_indx ON chunk_chunkfeat(freq);


CREATE TABLE IF NOT EXISTS batch(
             chunk_id INTEGER NOT NULL,
       	     name TEXT NOT NULL,
	     UNIQUE(chunk_id, name),
             FOREIGN KEY (chunk_id) REFERENCES chunk(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS batch_id_name ON batch(chunk_id, name);
CREATE INDEX IF NOT EXISTS batch_name ON batch(name);
CREATE INDEX IF NOT EXISTS batch_chunkid ON batch(chunk_id);

CREATE TABLE IF NOT EXISTS script(
             chunk_id INTEGER NOT NULL,
       	     name TEXT NOT NULL,
	     UNIQUE(chunk_id, name),
             FOREIGN KEY (chunk_id) REFERENCES chunk(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS script_id_name ON script(chunk_id, name);
CREATE INDEX IF NOT EXISTS script_name ON script(name);
CREATE INDEX IF NOT EXISTS script_chunkid ON script(chunk_id);

-- Example query for generating an ordered frequency list (takes some time):
-- select chunk_chunkfeat.chunkfeat_id, count(*) from chunk_chunkfeat, chunkfeat where chunkfeat.name = "word" and chunkfeat.id = chunk_chunkfeat.chunkfeat_id group by chunk_chunkfeat.chunkfeat_id order by count(chunk_chunkfeat.chunk_id) DESC;
--- Inserting into table (takes some time):
--- insert into wordfreq (chunkfeat_id, freq) select chunk_chunkfeat.chunkfeat_id, count(*) from chunk_chunkfeat, chunkfeat where chunkfeat.name = 'word' and chunkfeat.id = chunk_chunkfeat.chunkfeat_id group by chunk_chunkfeat.chunkfeat_id order by count(chunk_chunkfeat.chunk_id) DESC; 
CREATE TABLE IF NOT EXISTS wordfreq(
       id INTEGER  NOT NULL PRIMARY KEY AUTOINCREMENT,
       chunkfeat_id INTEGER NOT NULL,
       freq INTEGER NOT NULL,
       foreign key(chunkfeat_id) REFERENCES chunkfeat(id) ON DELETE CASCADE
       );


CREATE INDEX IF NOT EXISTS wordfreq_id_chunkfeat_id_indx ON wordfreq(id, chunkfeat_id);
CREATE INDEX IF NOT EXISTS wordfreq_freq_chunkfeat_id_indx ON wordfreq(freq, chunkfeat_id);
CREATE INDEX IF NOT EXISTS wordfreq_freq_id_indx ON wordfreq(freq);
CREATE INDEX IF NOT EXISTS wordfreq_id_chunkfeat_id_indx_2 ON wordfreq(chunkfeat_id);

----- VIEWS

CREATE VIEW IF NOT EXISTS chunk_wordcount_view AS SELECT chunk.*, chunk_chunkfeat.freq freq FROM  chunk, chunkfeat, chunk_chunkfeat WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunk_chunkfeat.chunkfeat_id = chunkfeat.id AND chunkfeat.name = 'count' AND chunkfeat.value = 'word_count';
-- TODO Pre-compute word count to make things snappier?
--CREATE VIEW IF NOT EXISTS chunk_wordcount_view AS SELECT chunk.*, COALESCE(SUM(NUMWORD.freq), 0) freq FROM chunk LEFT JOIN (SELECT chunk_chunkfeat.freq, c.id FROM chunkfeat, chunk_chunkfeat, chunk AS c WHERE chunkfeat.name = 'word' AND chunkfeat.id = chunk_chunkfeat.chunkfeat_id AND chunk_chunkfeat.chunk_id = c.id) NUMWORD ON chunk.id = NUMWORD.id GROUP BY chunk.id;



CREATE VIEW IF NOT EXISTS chunk_lowestwordfreqcount_view AS SELECT chunk.*, chunk_chunkfeat.freq freq FROM  chunk, chunkfeat, chunk_chunkfeat WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunk_chunkfeat.chunkfeat_id = chunkfeat.id AND chunkfeat.name = 'count' AND chunkfeat.value = 'lowest_word_freq';


-- CREATE VIEW IF NOT EXISTS chunk_commacount_view AS SELECT chunk.*, chunk_chunkfeat.freq freq FROM chunk, chunkfeat, chunk_chunkfeat WHERE chunk.id = chunk_chunkfeat.chunk_id AND chunkfeat.id = chunk_chunkfeat.chunkfeat_id AND chunkfeat.name = 'punct' AND chunkfeat.value = ',';
CREATE VIEW IF NOT EXISTS chunk_commacount_view AS SELECT chunk.*, COALESCE(SUM(NUMCOMMA.freq), 0) freq FROM chunk LEFT JOIN (SELECT chunk_chunkfeat.freq, c.id FROM chunkfeat, chunk_chunkfeat, chunk AS c WHERE chunkfeat.name = 'punct' AND chunkfeat.value = ',' AND chunkfeat.id = chunk_chunkfeat.chunkfeat_id AND chunk_chunkfeat.chunk_id = c.id) NUMCOMMA ON chunk.id = NUMCOMMA.id GROUP BY chunk.id;


CREATE VIEW IF NOT EXISTS chunk_digitcount_view AS SELECT chunk.*, COALESCE(SUM(NUMDIGIT.freq), 0) freq from chunk left join (SELECT chunk_chunkfeat.freq, c.id from chunkfeat, chunk_chunkfeat, chunk as c where chunkfeat.name = 'digit' and chunkfeat.id = chunk_chunkfeat.chunkfeat_id and chunk_chunkfeat.chunk_id = c.id) NUMDIGIT on chunk.id = NUMDIGIT.id GROUP BY chunk.id;
