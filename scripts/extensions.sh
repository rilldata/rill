#/bin/bash

DUCKDB_VERSION="v0.9.1"
DUCKDB_EXT="json.duckdb_extension icu.duckdb_extension parquet.duckdb_extension httpfs.duckdb_extension sqlite_scanner.duckdb_extension"
DUCKDB_EXT_URL="http://extensions.duckdb.org/$DUCKDB_VERSION/linux_amd64"

mkdir -p $HOME/.duckdb/extensions/$DUCKDB_VERSION/linux_amd64/
cd $HOME/.duckdb/extensions/$DUCKDB_VERSION/linux_amd64/

for d in ${DUCKDB_EXT}; do {
  echo "Installing duckdb extension: $d"
  wget -O $d.gz ${DUCKDB_EXT_URL}/$d.gz
  gzip -d $d.gz
} done
