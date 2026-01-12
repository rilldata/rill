# Resource YAML Training Data

Scrapes real-world Rill project resource YAML files from Rill Cloud for use as LLM input/training data.
Saves the output to `./scripts/resource-yaml-training-data/output/<type.txt>`.

## What it does

1. Uses `rill sudo project dump-resources --include-files` to fetch all resources of each type from Rill Cloud
2. Saves the raw JSON dumps to `output/<type>.json`
3. Formats the raw JSON dumps into a unified file of original YAML resources at `output/<type>.txt`

## Prerequisites

- Rill CLI installed and authenticated with admin access (`rill sudo` permissions)

## Usage

```bash
uv run ./scripts/resource-yaml-training-data/build.py
```
