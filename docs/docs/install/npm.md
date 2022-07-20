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
  
## Install or update with npm
You can use `npm` to globally install Rill Developer on your local computer. This will give you access to the [CLI](../cli.md) to start the server.

```
npm install -g @rilldata/rill
```

Note: this install command involves compiling DuckDB which can be time consuming to complete (it may take approximately five minutes or more, depending on your machine). Please be patient!



Rill Developer will be evolving quickly. If you want an updated version, you can pull in the latest changes and rebuild the application by running the same command, `npm install -g @rilldata/rill`. Once you have reinstalled the application you can restart your project to see the new experience.

## 404 errors
If you have just installed the application and are trying to see the User Interface at http://localhost:8080/ but see a 404 error, it is possible that npm install is taking longer than 5 minutes to build the application and you need to wait for the build to complete. Please wait an additional 5 minutes and try again.
