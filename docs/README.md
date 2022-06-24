# `docs/`

This folder contains docs for Rill, generated using [Docusaurus](https://docusaurus.io/) and deployed to [https://rilldata.github.io/rill-monorepo-prototype/](https://rilldata.github.io/rill-monorepo-prototype/).

## Local development

To start the docs server with hot reloading, run the following command from the _root_ of the repo:

```
npm run dev -w docs
```

## Preview production build

To run a full build and production preview of the docs, run the following commands from the _root_ of the repo:

```
npm run build -w docs
npm run preview -w docs
```

## Deploying the docs

The docs site is deployed to Github Pages using Github Actions.

## OpenAPI

The docs automatically generates an API reference based on our OpenAPI schema. We use [rohit-gohri/redocusaurus](https://github.com/rohit-gohri/redocusaurus) to add OpenAPI support to Docusaurus, which is a wrapper for [Redoc](https://github.com/Redocly/redoc). We considered also considered [cloud-annotations/docusaurus-openapi](https://github.com/cloud-annotations/docusaurus-openapi) and [PaloAltoNetworks/docusaurus-openapi-docs](https://github.com/PaloAltoNetworks/docusaurus-openapi-docs), but both embed Monaco, which heavily increases the bundle size.
