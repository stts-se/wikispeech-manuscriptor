# load_db

`load_db` is a CLI for loading a text corpus into a manuscript database.

Usage:

      go run cmd/load_db/main.go <options> <db file> <featcatdir> <WikiExtractor.py output files>

where `featcatdir` is the directory in which feature category/domain files reside. This repository contains a set of domain files, located in the `feat_data` folder: Swedish words for sports, weather, common names, etc. More information can be found in the documentation <a href="/doc/manuscript_tool.pdf">manuscript_tool.pdf</a> (Swedish only).

The above steps takes a lot of time and will eventually create a huge
database file. The database becomes very large, since for every
sentence in the corpus, a large amount of features and relations are added to
the database.
