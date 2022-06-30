#!/usr/bin/env bash

if [ -z ${PROJECT_BASE} ]; then
  PROJECT_BASE=".."
fi

echo "Initializing the project example project ${PROJECT_BASE}/rill-developer-example ..."
npm run cli --silent -- init --project ${PROJECT_BASE}/rill-developer-example

echo "Downloading dataset for example project..."
curl -s http://pkg.rilldata.com/rill-developer-example/data/flightlist.zip --output ${PROJECT_BASE}/rill-developer-example/flightlist.zip
unzip ${PROJECT_BASE}/rill-developer-example/flightlist.zip -d ${PROJECT_BASE}/rill-developer-example/

echo "Importing example dataset into the project..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/data/flightlist_2022_02.csv --project ${PROJECT_BASE}/rill-developer-example

# start the modeler
npm run cli --silent -- start --project ${PROJECT_BASE}/rill-developer-example
