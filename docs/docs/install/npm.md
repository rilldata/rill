---
description: You can use `npm` to globally install Rill Developer on your local computer. This will give you access to the CLI to start the server.
---

# Using npm

## Prerequisites

Nodejs version 16 installed locally: https://nodejs.org/en/download/. Check your version of Node:

```
node -v
```

On Ubuntu, you'll also need to make sure you have `g++` installed in order to compile DuckDB from source during the installation steps below (please note that compiling DuckDB may take a while):

```
sudo apt install g++
```
  
## Install
You can use `npm` to globally install Rill Developer on your local computer. 
```
npm install -g @rilldata/rill
```
Once installed, use the [CLI](../cli.md) to quick start a new project.

Note: this install command involves compiling DuckDB which can be time consuming to complete (it may take approximately five minutes or more, depending on your machine). Please be patient!

## Updates
Rill Developer will be evolving quickly. If you want an updated version, you can pull in the latest changes and rebuild the application by running the same command, `npm install -g @rilldata/rill`. Once you have reinstalled the application you can restart your project to see the new experience.

