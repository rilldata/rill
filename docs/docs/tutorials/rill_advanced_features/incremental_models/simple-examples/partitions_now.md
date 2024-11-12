---
title: "Partitions Models"
description:  "Start with basics"
sidebar_label: "Basic Partition Models"
sidebar_position: 2
---

Partitioned models are a special state of incremental models. Partitioned models are defined by defining a partition in the `partition` key.
Let's look at the following example.

```yaml
type: model

partitions:
  sql: SELECT range AS num FROM range(0,10)
sql: SELECT {{ .partition.num }} AS num, now() AS inserted_on
```

In this simple example, we set up 10 splits [range(0,10)] that have a single row with the same now() function as defined earlier. To confirm this we can run the following:

```bash
rill project partitions partitions_range --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-11-12T20:12:54Z   103ms            
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-11-12T20:12:54Z   0s               
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-11-12T20:12:54Z   0s               
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-11-12T20:12:54Z   0s               
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-11-12T20:12:54Z   0s               
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-11-12T20:12:54Z   0s               
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-11-12T20:12:54Z   1ms              
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-11-12T20:12:54Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-11-12T20:12:54Z   0s               
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-11-12T20:12:54Z   0s                  
  ```

If you try to refresh a single split, you'll receive the following error:
  
```bash
rill project refresh --model partitions_range --partition ff7416f774dfb086006d0b4696c214e1 --local          
Error: can't refresh partitions on model "partitions_range" because it is not incremental

```

## Incremental Partitioned Model
Bringing both concepts together, we can create a incremental partitioned model.

```yaml
type: model

splits:
  sql: SELECT range AS num FROM range(0,10)
sql: SELECT {{ .split.num }} AS num, now() AS inserted_on
incremental: true

output:
  incremental_strategy: merge
  unique_key: [num]
```

Similarily to the above, let's run `rill project partitions <model_name> --local` to get the key_ids.

```bash
rill project partitions partitions_range --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-11-12T20:12:54Z   103ms            
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-11-12T20:12:54Z   0s               
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-11-12T20:12:54Z   0s               
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-11-12T20:12:54Z   0s               
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-11-12T20:12:54Z   0s               
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-11-12T20:12:54Z   0s               
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-11-12T20:12:54Z   1ms              
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-11-12T20:12:54Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-11-12T20:12:54Z   0s               
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-11-12T20:12:54Z   0s          
```

Using the above information, we'll refresh the top split.

```bash
rill project refresh --model partitions_range_incremental --partition ff7416f774dfb086006d0b4696c214e1 --local          
Refresh initiated. Check the project logs for status updates.
```

Then, rerun the splits command to see that the EXECUTED ON columns has been updated.
```bash
rill project partitions partitions_range_incremental --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-11-12T20:18:55Z   3ms              
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-11-12T20:17:38Z   1ms              
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-11-12T20:17:38Z   1ms              
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-11-12T20:17:38Z   1ms              
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-11-12T20:17:38Z   1ms              
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-11-12T20:17:38Z   0s               
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-11-12T20:17:38Z   1ms              
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-11-12T20:17:38Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-11-12T20:17:38Z   0s               
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-11-12T20:17:38Z   0s       
```


The above is a static model. The split is defined by a set range(0,10) so there is no reason for us to put a automated refresh on this model. However, real data is likely not static and will require some sort of refresh when you push to Rill Cloud.

Now that we've gone over the basics, let's take a look at a more realistic example, our ClickHouse project.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />