#!/usr/bin/env bash

if [ ! -d ${PROJECT_BASE}/${PROJECT} ]; then
  echo "Initializing the project ${PROJECT}..."
  rill-developer init --project ${PROJECT_BASE}/${PROJECT}
fi

if [ "${PROJECT}" == "rill-developer-example" ] && [ ! -d ${PROJECT_BASE}/${PROJECT}/data ]; then
  echo "Downloading dataset for example project..."

  mkdir ${PROJECT_BASE}/${PROJECT}/data
  curl --progress-bar http://pkg.rilldata.com/rill-developer-example/data/aircraftDatabase-2022-02.csv --output ${PROJECT_BASE}/${PROJECT}/data/aircraftDatabase-2022-02.csv

  echo "Importing example dataset into the project..."
  rill-developer import-table ${PROJECT_BASE}/${PROJECT}/data/aircraftDatabase-2022-02.csv --project ${PROJECT_BASE}/${PROJECT}
else
  echo "Please refer to README.md for importing datasets."
fi

rill-developer start --project ${PROJECT_BASE}/${PROJECT}
