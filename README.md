Thank you for trying the Rill Data Modeler tech preview! This application is extremely alpha. 

We want to hear from you if you have any questions or ideas to share. You can file an issue directly in this repository or reach us in our Rill Community Slack at  https://bit.ly/35ijZG4.

## Prerequisites
Nodejs version 16+ installed locally: https://nodejs.org/en/download/. Check your version of Node:
```
node -v
```

## Install Locally
Change directories to the Rill Data Modeler Prototype
```
cd /path/to/data-modeler-prototype
```
Run npm to install dependencies and build the application. This will take ~5 minutes to complete.
```
npm install
npm run build
```
## Initialize your Project
Initialize your project in the data-modeler-prototype directory.
```
npm run cli -- init
```
Import datasets of interest into the Rill Data Modeler's duckDB database to make them available. We currently support .parquet and .csv. If you are looking for a public dataset to help you get started try the OpenSky Network datset at https://zenodo.org/record/6325961#.YjDFvhDMI0Q.
```
npm run cli -- import-table /path/to/data_1.parquet
npm run cli -- import-table /path/to/data_2.csv
```
## Start your Project
Start the User Interface to see your imported tables and revisit models you have created.
```
npm run cli -- start
```
The Data Modeler UI will be available at http://localhost:8080

## Data Modeler SQL Dialect
The Data Modeler is powered by duckDB. Please visit their documentation for insight into their dialect of SQL to facilitate data modeling. https://duckdb.org/docs/sql/introduction

## Updating the Data Modeler
The data modeler will be evolving quickly! If you want an updated version of the Data Modeler you can pull in the latest version of main and rebuild the application. Once you have rebuilt the application you can restart your project to see the new experience.
```
git pull origin main
npm run build
npm run cli -- start
```
## Helpful Hints
You can specify a new project folder by including the --project option.
```
npm run cli -- init --project /path/to/a/new/project
npm run cli -- import-table /path/to/data_1.parquet --project /path/to/a/new/project
npm run cli -- start --project /path/to/a/new/project
```
By default the table name will be a sanitized version of the dataset name. You can specify a specific name using the --name option.
```
npm run cli -- import-table  /path/to/data_1.parquet --name my_dataset
```
If you have a dataset that is delimited by a character other than a comma, you can use the --delimiter option.
```
npm run cli -- import-table /path/to/data_3.txt --delimiter "|"
```

# Application Developers
If you are a developer helping us build the application, please visit our [DEVELOPER-GUIDE.md](https://github.com/gorillio/data-modeler-prototype/blob/main/DEVELOPER-GUIDE.md).

