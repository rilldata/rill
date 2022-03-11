This is prototype-quality code, subject to radical change as we figure out what we need to build. Best of luck!

# Getting started

Run `npm install` to install all the dependencies and compile duckdb and other packages. This can take a long time to finish (~5mins).

Run `npm run server` to start the backend server.<br>
Run `npm run dev` to start the UI dev server. UI will be available on http://localhost:3000


# CLI

Interacting with a project currently is only supported through a cli.

### Developer CLI Usage

During development use an alias for `data-modeler` to `npm run cli-dev --`.
This will ensure that the same signature below in the cli docs can be used, but run the local code.

Adding an alias entry in .zshrc or .bashrc would be a good idea.
```
alias data-modeler="npm run cli-dev --"
```

`data-modeler-dev` can be used instead to have both the globally installed data-modeler and local development separate.
Just use `data-modeler-dev` in the below commands instead of `data-modeler`.

### End User CLI Usage

Install the package globally to access the cli.
```
npm i data-modeler -g
```
Run with sudo if `Error: EACCES: permission denied` is thrown with the above install.
```
sudo npm i data-modeler -g
```

### Creating a project

```
# init in current directory
data-modeler init
```
```
# init in /path/to/project directory
# directory will be created if it doesnt exist
data-modeler init --project /path/to/project
```

Note: This is not explicitly necessary.
Running the other commands on a non-existing directory or a fresh directory will automatically initialize the project.

### Importing a table from a file
```
# import table from /path/to/table/file into project under /path/to/project
data-modeler import-table /path/to/table/file --project /path/to/project

# Optionally pass a delimiter to override auto detected delimiter by duckdb.
# Only applies to a csv file
data-modeler import-table /path/to/table/csvfile --project /path/to/project --delimiter "|"
```
`--project` is optional. Will default to current directory if not specified.

Table name can be customisable using `--name` argument. By default, it uses file name without extension for table name.

**File types currently supported:**
 - .parquet
 - .csv
 - .tsv

### Starting the UI
```
# build the UI so that the server can server the built static files
# This is only needed for running in developer mode.
npm run build
# start the UI using info from project under /path/to/project
data-modeler start --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.


# Local testing

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
