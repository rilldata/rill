# Installing
Download and install nodejs 16+ from https://nodejs.org/en/download/.

Run the following script from checkout directory to install the command globally. This will take about 5mins to finish when run for the first time.

```
npm run install-and-build

```

# Developer Guide

## Getting started
Run `npm install` to install all the dependencies and compile duckdb and other packages. This can take a long time to finish (~5mins).<br>

Run `npm run build` to build the application.
 
## Starting a dev server
Run `npm run dev` to start the UI and backend dev servers. UI will be available on http://localhost:3000

# developer CLI
Initializing a project, adding datasets as sources, and starting a project are currently supported through our [CLI](https://github.com/rilldata/rill-developer/blob/main/docs/cli.md).

## Creating a project
Initialize in the current directory.
```
npm run cli --silent -- init
```

init in /path/to/project directory. The directory will be created if it doesnt exist.
```
npm run cli --silent -- init --project /path/to/project
```
*Note: This is not explicitly necessary. Running the other commands on a non-existing directory or a fresh directory will automatically initialize the project.*

Data modeler UI will be available at http://localhost:8080


## Starting the UI
Start the UI using info from project under /path/to/project.
```
npm run cli --silent -- start --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.


## Importing a source from a file
import source from /path/to/source/file into project under /path/to/project
```
npm run cli --silent -- import-source /path/to/source/file --project /path/to/project
```

Optionally pass a delimiter to override auto detected delimiter by duckdb.  Only applies to a csv file.
```
npm run cli --silent -- import-source /path/to/source/csvfile --project /path/to/project --delimiter "|"
```
`--project` is optional. Will default to current directory if not specified.
  
Source name can be customisable using `--name` argument. By default, it uses file name without extension for source name.

**File types currently supported:**
- .parquet

- .csv

- .tsv

## Dropping a source
Drop a source 'sourceToDrop' from project under /path/to/project
```
npm run cli --silent -- drop-source sourceToDrop --project /path/to/project
```
`--project` is optional. Will default to current directory if not specified.

