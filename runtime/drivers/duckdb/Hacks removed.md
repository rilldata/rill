Hacks removed. 

1. Removing `.tmp` and `.wal` directory. No longer works since the main.db no longer ingest anything.

Features removed.

1. String to `enum` conversion.
2. No `tx=true` queries since writes now happen on a different handle.