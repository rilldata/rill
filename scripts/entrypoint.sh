#!/usr/bin/env bash

if [ ! -z "${PROJECT_BASE}" ] && [ ! -z "${PROJECT}" ] && [ ! -d ${PROJECT_BASE}/${PROJECT} ]; then
  echo "Initializing the project ${PROJECT}..."
  rill-developer init --project ${PROJECT_BASE}/${PROJECT}
fi

if [ ! -z "${INIT_SCRIPT}" ] && [ -f ${INIT_SCRIPT} ]; then
 echo "Found init script at ${INIT_SCRIPT}..."
 source ${INIT_SCRIPT}
fi

echo "Starting Rill Developer at project ${PROJECT_BASE}/${PROJECT}..."
rill-developer start --project ${PROJECT_BASE}/${PROJECT}
