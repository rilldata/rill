---
title: Gemini
description: Use Gemini for AI features with your own API key
sidebar_label: Gemini
sidebar_position: 20
---

Rill supports connecting to Google's Gemini API using your own API key and configuration parameters for AI-powered coding and data analysis features.

## Steps

1. Obtain an API key from the [Google AI Studio](https://aistudio.google.com/apikey).
2. Add the API key as an environment variable in your Rill project called `gemini_api_key`.
    - See [Configure Credentials](/developers/build/connectors/credentials) for details.
    - Remember to sync the updated environment variables with `rill env pull` or `rill env push`.
3. Create a Gemini connector YAML file in your Rill project:
   ```yaml
   # connectors/gemini.yaml
   type: connector
   driver: gemini
   api_key: "{{ .env.gemini_api_key }}"
   model: "gemini-3-flash-preview"
   ```
4. Configure `gemini` as the project's default AI connector in `rill.yaml`:
   ```yaml
   # rill.yaml
   # ...
   ai_connector: gemini
   ```
