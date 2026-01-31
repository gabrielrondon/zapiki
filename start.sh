#!/bin/sh

# Railway Start Script
# Determines which service to start based on RAILWAY_SERVICE_NAME

if [ "$RAILWAY_SERVICE_NAME" = "zapiki-worker" ]; then
    echo "Starting Zapiki Worker..."
    exec ./zapiki-worker
else
    echo "Starting Zapiki API..."
    exec ./zapiki-api
fi
