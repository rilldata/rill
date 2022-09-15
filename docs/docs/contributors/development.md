---
Description: How to get started contributing to Rill Developer.
---

# Developer guide

## Installing
Download and install nodejs 16+ from https://nodejs.org/en/download/.

Run the following script from checkout directory to install the command globally. This will take about 5mins to finish when run for the first time.

```
npm run install-and-build
```

## Getting started
Run `npm install` to install all the dependencies and compile duckdb and other packages. This can take a long time to finish (~5mins).

Run `npm run build` to build the application.
 
## Starting a dev server
Run `npm run dev` to start the UI and backend dev servers. UI will be available on http://localhost:3000/

## Developer CLI
Initializing a project, adding datasets as sources, and starting a project are currently supported through our [CLI](cli).

## Creating a project

Set the following enviornment variable `RILL_IS_DEV=true` for marking the project as a dev project.

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
Import source from /path/to/source/file into project under /path/to/project
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
