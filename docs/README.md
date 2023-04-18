# `docs/`
[![Netlify Status](https://api.netlify.com/api/v1/badges/23baf08e-2d3e-44db-8bd4-938e54467a29/deploy-status)](https://app.netlify.com/sites/rill-developer/deploys)

This folder contains docs for Rill, generated using [Docusaurus](https://docusaurus.io/) and deployed to [https://docs.rilldata.com](https://docs.rilldata.com).

## Install packages

```
npm install
```

## Local development

To start the docs server with hot reloading, run the following command from the _root/docs_ folder of the repo:

```
npm run dev
```

## Preview production build

To run a full build and production preview of the docs, run the following commands from the _root/docs_ folder of the repo:

```
npm run build
npm run preview
```

## Deploying the docs

The docs site is deployed via Netlify.

## Creating a new version

```
npm run docusaurus docs:version <version> 
```
