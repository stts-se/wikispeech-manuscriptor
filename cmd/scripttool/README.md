# scripttool

`scripttool` is a CLI for manipulating a script database created according to instructions above. You create batches by filtering sentences in the database, and from these batches, you can create output manuscripts. You can also retreive information about the database, such as list existing batches/scripts, print db statistics, etc. For full usage and documentation, please invoke

      go run cmd/scripttool/main.go

Usage:

      scripttool <Sqlite3 manuscript db> <command> <args>
