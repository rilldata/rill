---
title: "Let's create a scheduled report"
description:  Let's create a scheduled report
sidebar_label: "Create a Scheduled Report"
sidebar_position: 2
---
## Pivot and Scheduled Reports

[Scheduling reports](https://docs.rilldata.com/explore/exports) is an important part of reporting. Via Scheduled Reports, you can setup a time-based export of [your pivot table](https://docs.rilldata.com/explore/filters/pivot) to specific individuals via email.


### Source refresh
Currently, Our datasets are currently static, we have not set up any [source refreshes](https://docs.rilldata.com/build/connect/source-refresh). Let's add the following lines to the source yaml files and push these changes to Rill Cloud.


```yaml
refresh:
  every: 24h
```
<details>
  <summary>Forgot how to push to GitHub? </summary>

   You can use the following commands to add, commit and push the changes via the CLI.
   ```
    $ git add sources 
    $ git commit -m "24hr refresh to sources"
        [main ef81ff1] 24hr refresh to sources
        2 files changed, 8 insertions(+), 2 deletions(-)
    $ git push origin main
        Enumerating objects: 9, done.
        Counting objects: 100% (9/9), done.
        Delta compression using up to 12 threads
        Compressing objects: 100% (5/5), done.
        Writing objects: 100% (5/5), 473 bytes | 473.00 KiB/s, done.
        Total 5 (delta 3), reused 0 (delta 0), pack-reused 0
        remote: Resolving deltas: 100% (3/3), completed with 3 local objects.
        To https://github.com/royendo/my-rill-tutorial.git
        07b0055..ef81ff1  main -> main

  Or, select the `redeploy` button from the UI, if you have not connected a GitHub repository.
   ```

</details>



Now, Rill will automaticaly refresh the datasets (our GCS connection) every 24 hours.  We are now ready to setup the scheduled report!


### Set up a scheduled report

Navigating back to the pivot table, we can select `Export`, `Create scheduled report`.

<img src = '/img/tutorials/205/scheduled-report.gif' class='rounded-gif' />
<br />


You will need to set the required fields as seen in the UI such as `title`, `frequency`, and `recipients`. Then, you can select `create` and view the scheduled export in the `Reports` tab of the project view.

![img](/img/tutorials/205/scheduled-report.png)

From the report view you can check the run history of the report, edit/delete the scheduled report, or manual run it. Only project or organization admins can manage reports.

:::note
You'll see in the UI whether the scheduled report was ran manually, Ad-hoc, or by the scheduler.
:::

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />