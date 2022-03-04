This is prototype-quality code, subject to radical change as we figure out
what we need to build. Best of luck!

To get started:

- get the DuckDB CLI from [https://duckdb.org/docs/installation/](https://duckdb.org/docs/installation/) and put it in `./server`
- run `npm install`
- run `npm run dev` to start the web UI dev server
- in a separate process, run `npm run server` to start the backend server

## CLI
We need to use ts-node-dev with tsconfig.node.json to get nodejs command working
```
# NOTE: -- after cli-dev is needed to pass the args to the cli
npm run cli-dev -- --help
```

### Creating a project
```
# init in current folder
npm run cli-dev -- init
```

```
# init in /path/to/project folder
npm run cli-dev -- init /path/to/project
```

### Importing a table from a file
```
# import table from /path/to/table/file into project under /path/to/project
npm run cli-dev -- import-table /path/to/table/file --project /path/to/project

# Optionally pass a delimiter to override auto detected delimiter by duckdb.
# Only applies to a csv file
npm run cli-dev -- import-table /path/to/table/csvfile --project /path/to/project --delimiter "|"
```
`--project` is optional. Will default to current directory if not specified.

**File types currently supported:**
 - .parquet
 - .csv
 - .tsv

### Starting the UI
```
# build the UI so that the server can server the built static files
npm run build
# start the UI using info from project under /path/to/project
npm run cli-dev -- start --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.

## Local testing
Generate local testing data using,
```
npm run generate-test-data
```
Will generate AdBids, AdImpressions and User data under /data
NOTE: this will only work by temporarily removing `"type": "module"` from package.js. This will be fixed in the future.

Run for test (Contains old tests that fail right now),
```
npm run test
```
