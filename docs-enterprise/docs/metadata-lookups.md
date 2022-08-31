---
title: "Metadata Lookups"
slug: "metadata-lookups"
excerpt: "Adding lookup tables for data enrichment"
hidden: false
createdAt: "2022-06-02T23:28:37.294Z"
updatedAt: "2022-06-02T23:28:37.294Z"
---
#Overview

Lookup tables are useful to reduce the size of your overall data set and to make sure that id/name combinations are always available with the latest data. Lookup examples would include:
- Campaign ID -> Campaign Name 
- Country Code -> Country 
- Account Owner -> Owner Name 

The best implementation path for lookups is to place the an output file (typically jsonl or csv) in an s3 or gcs bucket and point Druid towards that location. 

#Adding Lookups

Within Druid, you can access Lookups from the main control page (either input the URL for our console or select the Druid logo on the top left to return to the homepage. From that page, select **Lookups* 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/04c44e6-lookup.png",
        "lookup.png",
        2323,
        797,
        "#afb1ba"
      ]
    }
  ]
}
[/block]
On the next page, select **Add Lookups* to bring up the entry screen
 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/97665bf-addlookup.png",
        "addlookup.png",
        2323,
        719,
        "#abadb5"
      ]
    }
  ]
}
[/block]
On the Lookup entry screen, the first key change is the type - changing from map to cachedNamespace
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/32ae6fd-cachednamespace.png",
        "cachednamespace.png",
        2319,
        1159,
        "#9b9da9"
      ]
    }
  ]
}
[/block]
Changing the Type setting will bring up the necessary options to configure the lookup. From here enter:

*URL string of the location in the URI prefix
*Change the parse format to the filetype
*Update the polling period depending on frequency of change
[block:callout]
{
  "type": "info",
  "title": "Multi-column Files",
  "body": "Lookup files can contain a single id to name mapping or multiple fields. When selecting the parse format (either JSON or CSV) you have the option to define the Key and Value fields that are returned."
}
[/block]

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/da01b94-lookupsetup.png",
        "lookupsetup.png",
        2319,
        1049,
        "#989ba9"
      ]
    }
  ]
}
[/block]

[block:callout]
{
  "type": "warning",
  "title": "Injective Lookups",
  "body": "If your lookup is 1:1 meaning each ID has only one value, changing the setting to Injective = True will improve performance during query time"
}
[/block]