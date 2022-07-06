# Local testing
The test suite uses pre-generated data. This command generates `csv` and `parquet` files for AdBids, AdImpressions and User under `/data` before running the tests:
  
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