---
title: "Metadata Lookups"
slug: "metadata-lookups"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Adding lookup tables for data enrichment"/>

## Overview

Lookup tables are useful to reduce the size of your overall data set and to make sure that id/name combinations are always available with the latest data. Lookup examples would include:
- Campaign ID -> Campaign Name 
- Country Code -> Country 
- Account Owner -> Owner Name 

The best implementation path for lookups is to place the an output file (typically jsonl or csv) in an s3 or gcs bucket and point Druid towards that location. 

## Adding Lookups

Within Druid, you can access Lookups from the main control page (either input the URL for our console or select the Druid logo on the top left to return to the homepage. From that page, select **Lookups* 
![](https://images.contentful.com/ve6smfzbifwz/3aT5HhfECIpwT1ij9Ygzc0/432cac5f7bcc505e84d0434b188ceb77/0c4c937-Parent_Dash.png)
On the next page, select **Add Lookups* to bring up the entry screen
 
![](https://images.contentful.com/ve6smfzbifwz/sdhd6C7CJeh7fC7GS96US/f034372995de11ba2e561eccc9578c92/97665bf-addlookup.png)
On the Lookup entry screen, the first key change is the type - changing from map to cachedNamespace
![](https://images.contentful.com/ve6smfzbifwz/30qy5hYKTLSDcCChxjpVNU/8ca611843a91e8cee7d54dbe072e3e2b/32ae6fd-cachednamespace.png)
Changing the Type setting will bring up the necessary options to configure the lookup. From here enter:

*URL string of the location in the URI prefix
*Change the parse format to the filetype
*Update the polling period depending on frequency of change
:::info Multi-column Files
Lookup files can contain a single id to name mapping or multiple fields. When selecting the parse format (either JSON or CSV) you have the option to define the Key and Value fields that are returned.
:::

![](https://images.contentful.com/ve6smfzbifwz/2tOEJFSTqsztUgzVxxHoJ6/a2a3a366f575646dd71017500192bc48/da01b94-lookupsetup.png)

:::caution Injective Lookups
If your lookup is 1:1 meaning each ID has only one value, changing the setting to Injective = True will improve performance during query time
:::