#install and initialize
npm run cli --silent -- init --project ../rill-developer-example
mkdir ../rill-developer-example/data
echo 'downloading datasets...'
#download dataset and it's readme
curl -s http://pkg.rilldata.com/rill-developer-example/data/aircraftDatabase-2022-02.csv --output ../rill-developer-example/data/flightlist_2022_02.csv.gz
gunzip ../rill-developer-example/data/flightlist_2022_02.csv.gz
curl -s http://pkg.rilldata.com/rill-developer-example/data/flightlist.zip --output ../rill-developer-example/data/flightlist_README.md

# import tables
npm run cli --silent -- import-table ../rill-developer-example/data/flightlist_2022_02.csv --project ../rill-developer-example

# start the modeler
npm run cli --silent -- start --project ../rill-developer-example
