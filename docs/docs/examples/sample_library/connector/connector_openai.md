---
title: OpenAI Connector Example YAML
tags:
- connector
- code
- complete_file
- openai
docs: https://docs.rilldata.com/connect/data-source/openai
hash: 39b588bb0b6f4cc78cff37d90137541051749ada38f5b15445dcfce4f952ef8e
---

```yaml
type: connector
driver: openai
api_key: '{{ .env.openai_api_key }}'
model: gpt-4.1
temperature: 1
```
