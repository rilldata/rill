---
title: "Share your dashboards via Public URL"
sidebar_label: "Sharing Dashboards Publicly"
hide_table_of_contents: false
tags:
    - Rill Cloud
---

There might be occasions that sharing a dashboard to a 'non-user' is required in your workflow. In order to accomdate such situations, you can send a public URL of your dashboard with set parameters and with a expiration date if required.


### Accessing your Dashboard in Rill Cloud
Once you have deployed your project to Rill Cloud, you should be able to access the project via the following URL:

[Go to Rill Cloud](https://ui.rilldata.com/)

You can select your project from the list and select your dashboard.

![img](/img/tutorials/201/rill-cloud-projects.png)

### Adding Filters and Creating the Public URL

When sharing to your end-user, it is likely that you will want to specific a specific amount of filters. Once set, the end-user cannot view or modify the set filters so they will only be allowed to view a set portion of the dasboard that you define. 


Once your dashboard is ready, you can create the public URL via the `Share` button.

![img](/img/tutorials/other/public-url/share-public-url.png)


### Managing Public URLS in Rill Cloud

**via UI** 

Public URL can also be managed via the Settings page in Rill Cloud. Please refer to the [administrators guide](https://docs.rilldata.com/tutorials/administration/project-maintanence#public-url-management)!



**via CLI**

Checking the public URL can be done by running the following:

```bash
rill public-url list --project <your_project>
  ID                                     DASHBOARD     FILTER                                CREATED BY              CREATED ON            LAST USED ON          EXPIRES ON  
 -------------------------------------- ------------- ------------------------------------- ----------------------- --------------------- --------------------- ------------ 
  3564c499-c8bd-4c1c-bab8-c33a218a5009   advanced_metrics_view_explore   (author_name IN ["Alexey Milovidov""])   roy.endo@rilldata.com   2024-09-30 09:34:41   2024-09-30 09:34:41               
  cab99113-d5a8-468d-981e-213e41c7d1bf   advanced_metrics_view_explore                                         roy.endo@rilldata.com   2024-09-30 09:29:26   2024-09-30 09:34:32               

NOTE: For security reasons, the actual URLs can't be displayed after creation.
```
As you can see, you can receive information on who created, what filters, when it expires, etc.

### Example Public URL

Feel free to take a look at the created public URLs and note the difference between a dashboard with a filter and without.

[Public URL without filter](https://ui.rilldata.com/rill_learn/my-rill-tutorial/-/share/rill_mgc_boCupdujFIo0I0DFL7yoO3bAGdbaSqXWUn5OXIlXL8VeDDTENARBPv)


[Public URL with Filter](https://ui.rilldata.com/rill_learn/my-rill-tutorial/-/share/rill_mgc_1iB5JJ7CPz4g59vTihTnlYOkvqn8mYivtNxSpnvgnhc1IY56Pi86hs)