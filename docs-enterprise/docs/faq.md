---
title: "❓ FAQ"
slug: "faq"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="General queries and common asks"/>

## Where do you see the biggest validation issues come from?

There are several situations where Rill may not process data. Events that can not be processed fall into several categories and will be handled in the following ways:

  * Events posted unsuccessfully will be returned with status code: 420 or 503.
  * Events posted successfully (status code: 201) but with a file error (e.g. End of File error) will stall the pipeline. 
  * Events posted successfully but with invalid elements (e.g. malformed timestamp) will be dropped, and at this time no method exists to notify of this error.
  * Valid files with valid but incorrect JSON (e.g. wrong field name or incorrect nesting) will drop the fields corresponding to the incorrectly formed JSON.
## How secure will my data be?

Our security team implements the very best practices, both in our approach to safeguarding your data and in remaining vigilant through ongoing threat assessments and proactive responses. RCC is available via Auth0 to leverage your existing security checks and authentication.

Rill's service is used by Fortune 500 companies and has passed multiple infosec questionnaires. We are SOC II Type 2 compliant with reports available upon request.
## How am I charged for these services?

Our pricing model is a usage-based model similar to most cloud providers.

Billing is based on: 
  * volume and/or processing of the data being ingested
  * type of data storage (hot, cold, archival memory) 
  * compute required for unusual query volume (above amounts included with your storage capacity) 

You can [review our pricing page for more information](https://www.rilldata.com/pricing).

Please contact [support@rilldata.com](mailto:support@rilldata.com) or your TAM for more details.
