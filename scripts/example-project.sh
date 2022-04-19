#install and initialize
npm run cli --silent -- init --project ../rill-developer-example

echo 'downloading datasets...'
#download dataset and it's readme
curl -s http://pkg.rilldata.com/rill-developer-example/data/flightlist.zip --output ../rill-developer-example/flightlist.zip
unzip ../rill-developer-example/flightlist.zip -d ../rill-developer-example/

# import tables
npm run cli --silent -- import-table ../rill-developer-example/data/flightlist_2022_02.csv --project ../rill-developer-example

# start the modeler
npm run cli --silent -- start --project ../rill-developer-example

