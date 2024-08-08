---
title: "Installing Rill Developer"
sidebar_label: "Installing Rill Developer"
sidebar_position: 2
hide_table_of_contents: false
---

## Installing Rill [Linux / MacOS]

Let’s install the Rill binary to your local machine.

You can follow the steps in <a href="https://docs.rilldata.com/" target="_blank"> our documentation</a> or run the following from the CLI:

- On MacOS, open the CLI by searching for “Terminal” in Spotlight.

<img src = '/img/tutorials/101/Terminal.gif' class='rounded-gif' />
<br />


Once this is open run the following:

```yaml
curl https://rill.sh | sh
rill start my-rill-clickhouse
```

<details>
  <summary>Windows Installation Instructions</summary>

  On Windows, you can search for "Command Prompt" (note that there are extra steps to get Rill running on Windows; please refer to the <a href="https://docs.rilldata.com/" target="_blank">documentation</a> for more details).
  
  ``` yaml
        wsl --install -d Ubuntu-22.04
  ```
  Once the installation completes, and you have logged into the Linux instance, you need to install the unzip package using the following lines: 

    ```yaml
    sudo apt-get update
    sudo apt-get install unzip
    ```

    Finally, you can install Rill!
        ``` yaml 
            curl https://rill.sh | sh

    ```

</details>

<details>
  <summary>Want to skip a step? </summary>

    Later on in the course, we will sync our project to a GitHub repository. If you want, you can go ahead and create a repo in Git, then run the install script in the cloned location locally to make deployment easier. 

    More details on deploying Rill via Git in our docs' <a href='https://docs.rilldata.com/deploy/existing-project/' target="_blank"> Deploy section</a>.

</details>

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />


