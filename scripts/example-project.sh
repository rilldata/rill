#install and initialize
npm run cli --silent -- init --project ../data-modeler-example
mkdir ../data-modeler-example/data
echo 'downloading datasets...'
#download dataset and it's readme
curl -s https://zenodo.org/record/6325961/files/flightlist_20220201_20220228.csv.gz --output ../data-modeler-example/data/flightlist_2022_02.csv.gz
gunzip ../data-modeler-example/data/flightlist_2022_02.csv.gz
curl -s https://zenodo.org/record/6325961/files/readme.md --output ../data-modeler-example/data/flightlist_README.md

# import tables
npm run cli --silent -- import-table ../data-modeler-example/data/flightlist_2022_02.csv --project ../data-modeler-example

# start the modeler
npm run cli --silent -- start --project ../data-modeler-example
open http://localhost:8080
