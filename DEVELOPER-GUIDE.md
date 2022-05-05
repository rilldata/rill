This is prototype-quality code, subject to radical change as we figure out what we need to build. Best of luck!

# CLI

Initializing a project, adding datasets as tables, and starting a project are currently only supported through our CLI.

### Installing

Download and install nodejs 16+ from https://nodejs.org/en/download/.

Run the following script from checkout directory to install the command globally:
```
# This will take about 5mins to finish when run for the first time.
npm run install-and-build
```

### Creating a project

```
# init in current directory
npm run cli --silent -- init
```
```
# init in /path/to/project directory
# directory will be created if it doesnt exist
npm run cli --silent -- init --project /path/to/project
# Data modeler UI will be available at http://localhost:8080
```

Note: This is not explicitly necessary.
Running the other commands on a non-existing directory or a fresh directory will automatically initialize the project.

### Importing a table from a file
```
# import table from /path/to/table/file into project under /path/to/project
npm run cli --silent -- import-table /path/to/table/file --project /path/to/project

# Optionally pass a delimiter to override auto detected delimiter by duckdb.
# Only applies to a csv file
npm run cli --silent -- import-table /path/to/table/csvfile --project /path/to/project --delimiter "|"
```
`--project` is optional. Will default to current directory if not specified.

Table name can be customisable using `--name` argument. By default, it uses file name without extension for table name.

**File types currently supported:**
 - .parquet
 - .csv
 - .tsv

### Starting the UI
```
# start the UI using info from project under /path/to/project
npm run cli --silent -- start --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.

### Dropping a table
```
# Drop a table 'tableToDrop' from project under /path/to/project
npm run cli --silent -- drop-table tableToDrop --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.

# Developer Guide

## Getting started

Run `npm install` to install all the dependencies and compile duckdb and other packages. This can take a long time to finish (~5mins).<br>
Run `npm build` to build the application.

## Starting a dev server

Run `npm run dev` to start the UI and backend dev servers. UI will be available on http://localhost:3000

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
