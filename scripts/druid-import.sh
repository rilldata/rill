#!/bin/bash
## Usage Sample
# scripts/druid-import.sh --druid=<druid_url> --user=<username> --pass=<password> --datasource=chalice_sample --project=<project>
rows=5000000
project='.'

while [ $# -gt 0 ]; do
  case "$1" in
    --druid=*)
      druid="${1#*=}"
      ;;
    --user=*)
      user="${1#*=}"
      ;;
    --pass=*)
      pass="${1#*=}"
      ;;
    --datasource=*)
      datasource="${1#*=}"
      ;;
    --rows=*)
      rows="${1#*=}"
      ;;
    --project=*)
      project="${1#*=}"
      ;;

    *)
      printf "***************************\n"
      printf "* Error: Invalid argument.*\n$1"
      printf "***************************\n"
      exit 1
  esac
  shift
done


query="{\"query\" : \"SELECT * FROM \\\"$datasource\\\" limit $rows\",\"resultFormat\" : \"csv\", \"header\" : true}"
echo "$query" > "/tmp/$datasource-query.json"
printf "Downloading data using query :  $query \n"

curl -XPOST -H'Content-Type: application/json' -u "$user:$pass" https://$druid/druid/v2/sql/ -d @/tmp/$datasource-query.json > /tmp/$datasource.csv
printf "Importing to Rill Developer project: $project \n"
npm run cli --silent -- import-source /tmp/$datasource.csv --project $project

