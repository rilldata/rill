# using docker
 Docker is a containerization platform that packages our application and all its dependencies together to make sure it works seamlessly in any environment. You can install Rill Developer using our [docker container](https://hub.docker.com/r/rilldata/rill-developer
).

## build and compose
Build rill-developer using docker compose.
```
docker compose build
```

Run rill-developer using docker compose.
```
docker compose up
```
By default, this docker image will create a project named `rill-developer-example` under `./projects`.  The application should be running at [http://localhost:8080/](http://localhost:8080/)

## create a new project
To create a new project, update `PROJECT` in docker-compose.yml.

Copy all of the data files you would like to import into `./projects/${PROJECT}/data/`

```
docker exec -it rill-developer /bin/bash
rill import-source ${PROJECT_BASE}/${PROJECT}/data/<fileName> \
--project ${PROJECT_BASE}/${PROJECT}
```
  