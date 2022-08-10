#!/usr/bin/env bash

if [ -z ${PROJECT_BASE} ]; then
  PROJECT_BASE=".."
fi

echo "Initializing the project example project ${PROJECT_BASE}/rill-developer-example ..."
npm run cli --silent -- init --project ${PROJECT_BASE}/rill-developer-example

echo "Downloading dataset for example project..."
curl -s http://pkg.rilldata.com/rill-developer-example/example-assets.zip --output ${PROJECT_BASE}/rill-developer-example/example-assets.zip
unzip ${PROJECT_BASE}/rill-developer-example/example-assets.zip -d ${PROJECT_BASE}/rill-developer-example/

echo "Importing example datasets into the project..."
echo "Adtech..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/adtech-ad-click/adtech-item-data.csv --project ${PROJECT_BASE}/rill-developer-example
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/adtech-ad-click/adtech-train.csv --project ${PROJECT_BASE}/rill-developer-example
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/adtech-ad-click/adtech-view-log.csv --project ${PROJECT_BASE}/rill-developer-example
echo "Crypto..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/crypto-bitcoin/crypto-bitstamp-usd.csv --project ${PROJECT_BASE}/rill-developer-example
echo "Ecommerce..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/ecomm-click-stream/e-shop-clothing.csv --project ${PROJECT_BASE}/rill-developer-example
echo "Global..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/global-landslide-catalog/global-landslide-catalog.csv --project ${PROJECT_BASE}/rill-developer-example
echo "Internet of Things..."
npm run cli --silent -- import-source ${PROJECT_BASE}/rill-developer-example/example-assets/data/iot-env-sensor/iot-telemetry-data.csv --project ${PROJECT_BASE}/rill-developer-example

mkdir  ${PROJECT_BASE}/rill-developer-example/data
mv -v ${PROJECT_BASE}/rill-developer-example/example-assets/data ${PROJECT_BASE}/rill-developer-example

echo "Importing example SQL transforamtions into the project..."
mv -v ${PROJECT_BASE}/rill-developer-example/example-assets/models/* ${PROJECT_BASE}/rill-developer-example/models

echo "Cleaning up the project..."
rm ${PROJECT_BASE}/rill-developer-example/models/model_1.sql
rm -rf ${PROJECT_BASE}/rill-developer-example/example-assets
rm -rf ${PROJECT_BASE}/rill-developer-example/__MACOSX
rm -rf ${PROJECT_BASE}/rill-developer-example/example-assets.zip

echo "Starting example..."
npm run cli --silent -- start --project ${PROJECT_BASE}/rill-developer-example
