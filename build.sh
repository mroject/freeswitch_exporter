#!/bin/bash

# Build the Docker image
docker build -t freeswitch_exporter .

# Run a temporary container based on the image
docker run -d --name temp_container freeswitch_exporter

# Copy the binary out of the container
docker cp temp_container:/freeswitch_exporter ./freeswitch_exporter

# Stop and remove the temporary container
docker rm -f temp_container
