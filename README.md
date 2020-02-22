# simple-service-registry
A simple API service registry in Go.

## Functional overview
This is a simple API registry to allow the self-registration and discovery of HTTP APIs.  It accepts POSTs of new service objects (with no authentication, so take that into account when deploying), and then makes a catalog of these services available for retrieval.

There are two storage backends implemented to maintain service state:
 - Map ('map' in the config): uses a basic Go map to maintain a fast, non-persistent in-memory store of services
 - Boltdb ('boltdb' in the config): uses the excellent [boltdb](https://github.com/boltdb/bolt) to maintain a persistent kv store of services

 Additionally, a polling layer can be enabled over the service store to hit the optionally provided healthcheck URLs of each service; if this is used, then the catalog will include useful information about service availability.

This is a simple, lightweight registry designed to satisfy basic catalog/discovery requirements - behaviour under high load and for very large numbers of services has not been tested.

## Building the binary
Run `make gobuild` from the base directory; the registry binary will be built.

## Building a docker image
With docker installed, run `make image` from the base directory.

## Running with docker
To test run with docker (removes all data when the container stops):

`docker run --rm -ti -p 8080:8080 barelyuseless/simple-service-registry:latest`

To run and persist data to a bind mount:

`docker run --rm -ti -p 8080:8080 -v $(pwd)/data:/data barelyuseless/simple-service-registry:latest`

## API Specification
Once running, check /spec for the swagger API specification.