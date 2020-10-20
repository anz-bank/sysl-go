# Template

sysl file to test template generation

## Prerequisites

- [Sysl v0.11.0 or later ](https://sysl.io/docs/install/)
- Go 1.13

## Building application Docker container image

To build a docker container image for the application using the template
Dockerfile, you first need to vendor all the dependencies. You can do so
with the command
```sh
go mod vendor
```

Then run:

```sh
docker build -t docker_template -f Dockerfile .
```

## Running the application in a Docker container

o run the application inside the container, you need to prepare an
application config file. Assuming you have an application config file
present as `config.yml`, the containerised application can be run as:

```sh
docker run --rm -t -p 8080:8080 --mount type=bind,source="$PWD"/config.yml,target=/app/config.yml,readonly docker_template:latest /bin/template /app/config.yml
```
