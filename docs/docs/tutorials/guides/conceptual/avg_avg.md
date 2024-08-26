---
title: "Average of Averages, aren't actually averages, why?"
sidebar_label: "Average of Averages"
sidebar_position: 10
hide_table_of_contents: false
---
To understand why this is an issue when using OLAP engines, we need to both understand why this occurs, and the mathematical reasoning behind it.

## What is an average of an average? Why is it not a total average?

Mathematically speaking, an average does not always equal an average of averages. Why?
Simply based on how mathematics works. In order for an average of average to be correct, the number of values in each group needs to be equal. Let's take the following example:

You have four numbers: `20`, `40`, `60` and `80`. Average these four numbers give us the value of `200`/`4` = `50`. 

Now let's try grouping them in a different way by taking the average of `20` by itself, and `40`, `60` and `80`, then averaging the averages.

`20`/ `1` = `20`

`180`/`3` = `60`

Taking the above and averaging results in `40`, not `50`.

`80` / `2` = `40`

In the same example, if we group them each by 2:

`60`/ `2` = `30`

`140` / `2` = `70`

This results in `50`, which mataches because the number of values in each grouping is equal.
`100` / `2` = `50`

## Why this matters in Rill
This matters in Rill because of how OLAP engines work. While each engine works slightly different, when managing the underlying data, there is no guarantee that your data is grouped into one when calculating averages, especially if you have large amounts of data. This is why when writing your model's SQL, we advise against certain calculations like AVG().

>insert AVG() screenshot in Model with a no-no

:::note Performance

:::

## What you can do to avoid this issue? 
Instead, after materializing the final output model table, you can create a measure using the AVG() function.

>insert screenshot of avg in measures or copy some YAML 

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />