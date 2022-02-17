This is prototype-quality code, subject to radical change as we figure out
what we need to build. Best of luck!

To get started:

- get the DuckDB CLI from [https://duckdb.org/docs/installation/](https://duckdb.org/docs/installation/) and put it in `./server`
- run `npm install`
- run `npm run dev` to start the dev server
- in a separate process, run `npm run server` to start the server

## CLI
We need to use ts-node-dev with tsconfig.node.json to get nodejs command working
```
npx ts-node-dev --project tsconfig.node.json -- src/cli/data-modeler-cli.ts --help
```

## Local testing data
Generate local testing data using,
```
npx ts-node-dev --project tsconfig.node.json -- test/data/generator/generate-data.ts
```
Will generate AdBids, AdImpressions and User data under /data
