---
title: Google Sheets
description: Connect to data in Google Sheets
sidebar_label: Google Sheets
sidebar_position: 20
---

### Google Sheets

Rill can read from any HTTP(S) URL endpoint that produces a valid data file in a supported format. For example, to bring in data from [Google Sheets](https://www.google.com/sheets/about/) as a CSV file directly into Rill as a source ([leveraging the direct download link syntax](https://www.highviewapps.com/blog/how-to-create-a-csv-or-excel-direct-download-link-in-google-sheets/)), you can create a `source_name.yaml` file in the `sources` directory of your Rill project with the following content. 

:::note Prerequisite 

The Google Sheet must be a public sheet with `Anyone with the link` enabled.

:::

```yaml
type: model
connector: "duckdb"
sql: "select * from read_csv_auto('https://docs.google.com/spreadsheets/d/<SPREADSHEET_ID>/export?format=csv&gid=<SHEET_ID>', normalize_names=True)"
```

:::note Updating the URL

Make sure to replace `SPREADSHEET_ID` and `SHEET_ID` with the ID of your spreadsheet and tab, respectively (which you can obtain from the URL when Google Sheets is open).

:::

<img src='/img/build/connectors/data-sources/googlesheets.png' class='rounded-gif' />
<br />

:::tip `gsheets` DuckDB Community Extension
In cases where setting the Google Sheet to `Anyone with the link` is not allowed, DuckDB has an extension that allows you to share the sheet to a Google service account. However, we have not yet implemented the DuckDB community extension in our deployment of DuckDB. Please contact us if you are interested in using this feature.
:::
