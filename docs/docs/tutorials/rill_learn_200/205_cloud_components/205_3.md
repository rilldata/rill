---
title: "A dashboard to share externally"
description:  How Shareable URLs work
sidebar_label: "Create a Public URL"
sidebar_position: 2
---

## A fixed view dashboard
There will be times that you will need to [share a dashboard externally]. In this case, you can create a temporary public URL with set filters, that are not modifable by the viewers.

<img src = '/img/tutorials/205/public-url.gif' class='rounded-gif' />
<br />


Depending on the time that you want to keep this URL active, you can set a expiration date. In the case that this option is disabled, the URL will be able to be access indefinitely. To manage a  public url, this is currently possible via the CLI.

### Via the CLI
```
rill public-url 

  list        List all public URLs
  create      Create a public URL
  delete      Delete a public URL
```
Example:
```
rill public-url list
  ID                                     DASHBOARD     FILTER                                CREATED BY              CREATED ON            LAST USED ON          EXPIRES ON           
 -------------------------------------- ------------- ------------------------------------- ----------------------- --------------------- --------------------- --------------------- 
  7b825298-734a-4d56-8957-da377aff99c6   dashboard_1   (author_name IN ["Robert Schulze"])   roy.endo@rilldata.com   2024-08-07 18:03:56   2024-08-07 18:03:56   2024-10-05 08:59:56  
  ```

### Via the UI
import ComingSoon from '@site/src/components/ComingSoon';

<ComingSoon />

<div class='contents_to_overlay'>

</div>




import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />
