#!/usr/bin/env bash

if [ ! -d ${PROJECT_BASE}/${PROJECT} ]; then
  echo "Initializing the project ${PROJECT}..."
  rill-developer init --project ${PROJECT_BASE}/${PROJECT}
fi

if [ "${PROJECT}" == "rill-developer-example" ] && [ ! -d ${PROJECT_BASE}/${PROJECT}/data ]; then
  echo "Downloading dataset for example project..."
  mkdir ${PROJECT_BASE}/${PROJECT}/data

  curl --progress-bar https://zenodo.org/record/6325961/files/flightlist_20220201_20220228.csv.gz --output ${PROJECT_BASE}/${PROJECT}/data/flightlist_2022_02.csv.gz
  curl -s https://zenodo.org/record/6325961/files/readme.md --output ${PROJECT_BASE}/${PROJECT}/data/flightlist_README.md
  gunzip ${PROJECT_BASE}/${PROJECT}/data/flightlist_2022_02.csv.gz

  echo "Importing example dataset into the project..."
  rill-developer import-table ${PROJECT_BASE}/${PROJECT}/data/flightlist_2022_02.csv --project ${PROJECT_BASE}/${PROJECT}
else
  echo "Please refer to README.md for importing datasets."
fi

rill-developer start --project ${PROJECT_BASE}/${PROJECT}
