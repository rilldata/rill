This is prototype-quality code, subject to radical change as we figure out
what we need to build. Best of luck!

To get started:

- get the DuckDB CLI from [https://duckdb.org/docs/installation/](https://duckdb.org/docs/installation/) and put it in `./server`
- run `npm install`
- run `npm run dev` to start the dev server
- in a separate process, run `node server` to start the server

## sql engine API

We will need to expose a few functions from an API:

- `createSourceProfile` – creates a "profile" of the source table(s)
- `createDestinationProfile` – creates a "profile" of the output dataframe

## order of operations on startup
- check the dbs in ./staging-databases for a databases.json file, which has an array of dbs with:
    - path to the parquet file
    - path to the duckdb stage
- if one of those doesn't exist, then we do something.

