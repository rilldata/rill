# Rill Templates

## Defining templates

The types for a template are written down in `types.ts`. These types are converted to JSON schema using `ts-json-schema-generator`

To generate the JSON schema for validation on the runtime, run

```
npm run generate:template-schema -w web-common
```