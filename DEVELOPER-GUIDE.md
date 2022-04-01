This is prototype-quality code, subject to radical change as we figure out what we need to build. Best of luck!

# Local Developer Setup

## Getting started

Run `npm install` to install all the dependencies and compile duckdb and other packages. This can take a long time to finish (~5mins).<br>
Run `npm build` to build the application.

## Starting a dev server

Run `npm run server` to start the backend server.<br>
Run `npm run dev` to start the UI dev server. UI will be available on http://localhost:3000

## Local testing

The test suite uses pre-generated data. Thus, you will need to run the following command before running the tests:
```
npm run generate-test-data
```
csv and parquet files for AdBids, AdImpressions and User datasets are generated under /data

Check test/generator/types for schema for AdBids, AdImpressions and User.

Run this command to run the test suite:
```
npm run test
```

Run individual test files by running jest directly:
```
npx jest /path/to/test/file
```

If you're working on the UI and want to make changes to UI tests, you can run 

```
npm run test:ui
```

The UI tests utilize [Playwright](https://github.com/microsoft/playwright/blob/main/LICENSE). Thus you can easily add common flags. For instance, if you need to run the visual / code debugger, run

```
PWDEBUG=1 npm run test:ui
```
