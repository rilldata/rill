#!/usr/bin/env bash

if [ ! -z "${INIT_SCRIPT}" ] && [ -f ${INIT_SCRIPT} ]; then
 echo "Found init script at ${INIT_SCRIPT}..."
 source ${INIT_SCRIPT}
fi

echo "Starting Rill Developer..."
/app/rill start
