# Rill Developer **_(tech preview)_**
Rill Developer makes it effortless to transform your datasets with SQL and create powerful, opinionated dashboards. Rill's principles:

- *feels good to use* – powered by Sveltekit & DuckDB = conversation-fast, not wait-ten-seconds-for-result-set fast
- *works with your local and remote datasets* – imports and exports Parquet and CSV
- *no more data analysis "side-quests"* – helps you build intuition about your dataset through automatic profiling
- *no "run query" button required* – responds to each keystroke by re-profiling the resulting dataset
- *radically simple dashboards* - thoughtful, opinionated defaults to help you quickly derive insights from your data

![Kapture 2022-07-21 at 15 34 45](https://user-images.githubusercontent.com/5587788/180313797-ef50ec6e-fc2d-4072-bb77-b2acf59205d7.gif "732257485")

## Installation

On macOS, we recommend installing `rill` using Brew:

```bash
brew install rilldata/tap/rill
```

On Linux, we recommend installing `rill` using our installation script:

```bash
curl -s https://cdn.rilldata.com/install.sh | bash
```

<!-- TODO: Add docs link here -->

## Quick start

You start Rill using the CLI. To start Rill in a new, empty project:

```bash
rill init --project my-project
cd my-project
rill start
```

You can also check out the Rill Developer repository and explore one of our example projects. For example:

```bash
git clone https://github.com/rilldata/rill-developer.git
cd rill-developer/examples/sf_props
rill start
```

## We want to hear from you

You can [file an issue](https://github.com/rilldata/rill-developer/issues/new/choose) directly in this repository or reach us in our [Discord channel](https://bit.ly/3unvA05). Please abide by the [Rill Community Policy](https://github.com/rilldata/rill-developer/blob/main/COMMUNITY-POLICY.md).

## Legal
By downloading and using our application you are agreeing to the [Privacy Policy](https://www.rilldata.com/legal/privacy) and [Rill Terms of Service](https://www.rilldata.com/legal/tos).
