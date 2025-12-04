# `docs/`

[![Netlify Status](https://api.netlify.com/api/v1/badges/23baf08e-2d3e-44db-8bd4-938e54467a29/deploy-status)](https://app.netlify.com/sites/rill-developer/deploys)

This folder contains docs for Rill, generated using [Docusaurus](https://docusaurus.io/) and deployed to [https://docs.rilldata.com](https://docs.rilldata.com).

## Building the docs

### Install packages

```
npm install
```

### Local development

To start the docs server with hot reloading, run the following command from the _root/docs_ folder of the repo:

```
npm run dev
```

### Preview production build

To run a full build and production preview of the docs, run the following commands from the _root/docs_ folder of the repo:

```
npm run build
npm run preview
```

## Deploying the docs

The docs site is deployed via Netlify.

## Generated docs

### CLI reference

The CLI reference docs in `docs/reference/cli` are auto-generated based on the CLI help text. To re-generate the docs, run the following command from the repository root:

```bash
make docs.generate
```

## Hiding content

Sometimes you want to merge documentation but hide it pending a future release, there are a few different short-hands available.  
For individual pages add `sidebar_class_name: hidden` to the top of the page and to hide an entire category add `className: hidden` to the `_category_.yml` file. Note that the pages are only hidden but still generated so you can still link to them.
