# Rill Developer **_(tech preview)_**
Rill Developer is a tool that makes it effortless to transform your datasets with SQL. It's not just a SQL GUI! Rill Developer follows a few guiding principles:

- _no more data analysis "side-quests"_ – helps you build intuition about your dataset through automatic profiling
- _no "run query" button required_ – responds to each keystroke by re-profiling the resulting dataset
- _works with your local datasets_ – imports and exports Parquet and CSV
- _feels good to use_ – powered by Sveltekit & DuckDB = conversation-fast, not wait-ten-seconds-for-result-set fast

It's best to show and not tell, so here's a little preview of Rill Developer:

![RillDeveloper](https://user-images.githubusercontent.com/5587788/160640657-2b68a230-9dcb-4236-a6c8-df5263c33443.gif)

## We want to hear from you if you have any questions or ideas to share
You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Rill discord](https://bit.ly/3unvA05) channel. Please abide by the [rill community policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).

## Pick an install option:
- [binary](https://docs.rilldata.com/install/binary) : download the most recent [assets binary](https://github.com/rilldata/rill-developer/releases).
- [npm](https://docs.rilldata.com/install/npm) : run  `npm install -g @rilldata/rill`
- [docker](https://docs.rilldata.com/install/docker) : download our [docker container](https://hub.docker.com/r/rilldata/rill-developer)

## Quick start a new project
You can create and augment your own projects in Rill Developer using the [CLI](https://docs.rilldata.com/cli). Every project starts by initializing the experience. Once initialized, you can ingest data into the project and start the UI.

```
rill init
rill import-source /path/to/data_1.parquet
rill start
```

## Try an example project
If you want to see several examples with data transformations you can install our example project.
```
rill init-example-project
```
This project imports 7 datasets and performs simple transfomrations to create 5 analytics ready resources. For more information on the datasets, please see the readme.md files in the `data` rill-developer-example directory after running the script.

## More information
See our [documentation](https://docs.rilldata.com) for more information.

## Legal
By downloading and using our application you are agreeing to the [Rill Terms of Service](https://www.rilldata.com/legal/tos) and [Privacy Policy](https://www.rilldata.com/legal/privacy).
