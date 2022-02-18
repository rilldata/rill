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

## Local testing data
Generate local testing data using,
```
npm run generate-test-data
```
Will generate AdBids, AdImpressions and User data under /data
NOTE: this will only work by temporarily removing `"type": "module"` from package.js. This will be fixed in the future.