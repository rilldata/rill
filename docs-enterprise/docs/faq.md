---
title: "❓ FAQ"
slug: "faq"
excerpt: "General queries and common asks"
hidden: false
createdAt: "2021-06-17T03:23:37.531Z"
updatedAt: "2022-07-13T20:15:51.316Z"
---
[block:api-header]
{
  "title": "Where do you see the biggest validation issues come from?"
}
[/block]
There are several situations where Rill may not process data. Events that can not be processed fall into several categories and will be handled in the following ways:

  * Events posted unsuccessfully will be returned with status code: 420 or 503.
  * Events posted successfully (status code: 201) but with a file error (e.g. End of File error) will stall the pipeline. 
  * Events posted successfully but with invalid elements (e.g. malformed timestamp) will be dropped, and at this time no method exists to notify of this error.
  * Valid files with valid but incorrect JSON (e.g. wrong field name or incorrect nesting) will drop the fields corresponding to the incorrectly formed JSON.
[block:api-header]
{
  "title": "How secure will my data be?"
}
[/block]
Our security team implements the very best practices, both in our approach to safeguarding your data and in remaining vigilant through ongoing threat assessments and proactive responses. RCC is available via Auth0 to leverage your existing security checks and authentication.

Rill's service is used by Fortune 500 companies and has passed multiple infosec questionnaires. We are SOC II Type 2 compliant with reports available upon request.
[block:api-header]
{
  "title": "How am I charged for these services?"
}
[/block]
Our pricing model is a usage-based model similar to most cloud providers.

Billing is based on: 
  * volume and/or processing of the data being ingested
  * type of data storage (hot, cold, archival memory) 
  * compute required for unusual query volume (above amounts included with your storage capacity) 

You can [review our pricing page for more information](https://www.rilldata.com/pricing).

Please contact [support@rilldata.com](mailto:support@rilldata.com) or your TAM for more details.