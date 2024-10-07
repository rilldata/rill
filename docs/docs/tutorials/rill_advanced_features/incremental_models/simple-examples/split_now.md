---
title: "Split Models"
description:  "Start with basics"
sidebar_label: "Basic Split Models"
sidebar_position: 2
---

As mentioned, we can define a split by adding the `splits:` key with some defining parameters. Splits are a special case of incremental model states. Let's look at the following example.

```yaml
type: model

splits:
  sql: SELECT range AS num FROM range(0,10)
sql: SELECT {{ .split.num }} AS num, now() AS inserted_on
```

In this simple example, we set up 10 splits [range(0,10)] that have a single row with the same now() function as defined earlier. To confirm this we can run the following:

```bash
rill project splits splits_range --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-09-18T02:32:01Z   145ms            
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-09-18T02:32:01Z   0s               
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-09-18T02:32:01Z   0s               
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-09-18T02:32:01Z   1ms              
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-09-18T02:32:01Z   1ms              
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-09-18T02:32:01Z   1ms              
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-09-18T02:32:01Z   0s               
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-09-18T02:32:01Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-09-18T02:32:01Z   1ms              
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-09-18T02:32:01Z   0s       
  ```

  If you try to refresh a single split, you'll receive the following error:
  
  ```bash
rill project refresh --model splits_range --split ff7416f774dfb086006d0b4696c214e1 --local
Error: can't refresh splits on model "splits_range" because it is not incremental
```

### Incremental Split Model
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

Similarily to the above, let's run `rill project splits <model_name>` to get the key_ids.

```bash
rill project splits splits_range_incremental --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-09-18T03:17:16Z   103ms            
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-09-18T03:17:16Z   1ms              
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-09-18T03:17:16Z   0s               
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-09-18T03:17:16Z   0s               
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-09-18T03:17:16Z   0s               
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-09-18T03:17:16Z   0s               
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-09-18T03:17:16Z   0s               
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-09-18T03:17:16Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-09-18T03:17:16Z   0s               
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-09-18T03:17:16Z   0s       
```

Using the above information, we'll refresh the top split.

```bash
rill project refresh --model splits_range_incremental --split ff7416f774dfb086006d0b4696c214e1 --local 
Refresh initiated. Check the project logs for status updates.
```

Then, rerun the splits command to see that the EXECUTED ON columns has been updated.
```bash
royendo@Roys-MacBook-Pro-2 modeling % rill project splits splits_range_incremental --local
  KEY (10)                           DATA        EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- ----------- ---------------------- --------- ------- 
  ff7416f774dfb086006d0b4696c214e1   {"num":0}   2024-09-18T03:17:58Z   1ms              
  69401118e166742864f35f1a77ffe07d   {"num":1}   2024-09-18T03:17:16Z   1ms              
  555ef019f87b5a57ec7b057476fe9d38   {"num":2}   2024-09-18T03:17:16Z   0s               
  09a5530e62c87e41848a02680f4422d4   {"num":3}   2024-09-18T03:17:16Z   0s               
  ecc2e2f9deb13509547ebcc2e5c55116   {"num":4}   2024-09-18T03:17:16Z   0s               
  1e3ddd76525fefd5ac9989d9b6c4727e   {"num":5}   2024-09-18T03:17:16Z   0s               
  25684ad5ef8f1965b597edeeb8004afa   {"num":6}   2024-09-18T03:17:16Z   0s               
  8142d250d75d0a20883333c00c3962d5   {"num":7}   2024-09-18T03:17:16Z   0s               
  0d6962a0746cb896ce87250808a50051   {"num":8}   2024-09-18T03:17:16Z   0s               
  727d91a916260837579d5e42ad696dd9   {"num":9}   2024-09-18T03:17:16Z   0s     
```


The above is a static model. The split is defined by a set range(0,10) so there is no reason for us to put a refresh key-pair on this. However, real data is likely not static and will require some sort of refresh when you push to Rill Cloud.

Let's take a look at a more realistic example, our ClickHouse project.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />