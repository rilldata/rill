---
title: "1. Launch Rill Developer"
sidebar_label: "1. Launch Rill Developer"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:MotherDuck
  - Tutorial
---

:::note prerequisites
You will need to [install Rill](https://docs.rilldata.com/home/install).

```bash
curl https://rill.sh | sh
```

You need a MotherDuck token to connect to your MotherDuck database. 
Check MotherDuck's documentation on [how to generate an access token](https://motherduck.com/docs/key-tasks/authenticating-and-connecting-to-motherduck/authenticating-to-motherduck/#authentication-using-an-access-token).


:::
## Start Rill Developer

```bash
rill start my-rill-motherduck
```

After running the command, Rill Developer should automatically open in your default browser. If not, you can access it via the following url:

```
localhost:9009
``` 

You should see the following webpage appear: 

<img src = '/img/tutorials/rill-basics/new-rill-project.png' class='rounded-gif' />
<br />

Let's go ahead and select `Start with an empty project`.

<details>
  <summary>Where am I in the terminal?</summary>
  
    You can use the `pwd` command to see which directory in the terminal you are. <br />
    If this is not where you'd like to make the directory use the `cd` command to change directories.

</details>



