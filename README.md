# JAM

[![JAM](https://serhiy.s3.eu-central-1.amazonaws.com/Github_repo/JAM/cover.png)](https://github.com/exelban/JAM)

Just Another Monitoring

## Description
JAM is a simple monitoring tool application. 
It allows you to monitor the status of your services and applications by sending HTTP requests to them and checking the response status code.
The main idea is to have a simple and easy-to-use monitoring tool with minimalistic and nice design.

For now it is in the development stage and has a lot of features to be implemented. Such as proper alerts, more monitoring options, events history and more.

## Features
- 90 days history
- groups of hosts
- alerts (in progress)
- events history (in progress)
- multiple databases support (in progress, only bolt and in-memory for now)

## Installation

Application is available as a Docker image. You can pull it from the Docker Hub or GitHub Registry:
- [exelban/jam:latest](https://hub.docker.com/r/exelban/jam)
- [ghcr.io/exelban/jam:latest](https://github.com/users/exelban/packages/container/package/jam)

Also you can build it from the source code or use the precompiled binaries. But docker is the easiest way to run the application. And it is recommended to use it.

### Docker
```bash
docker run -d -v ./jam.yaml:/app/config.yaml exelban/jam:latest
```

### Docker Compose
```yaml
services:
  jam:
    image: exelban/jam:latest
    container_name: jam
    restart: unless-stopped
    volumes:
      - ./jam.yaml:/app/config.yaml
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    healthcheck:
      test: "curl -f http://localhost:8822/healthz || exit 1"
      interval: 10s
      timeout: 10s
      retries: 3
      start_period: 3s
```

### Precompiled binaries
You can download the precompiled binaries from the [releases](https://github.com/exelban/JAM/releases) page.

### Build from source
To build the application from the source code you need to have [Go](https://go.dev/doc/install) installed on your machine.

```bash
git clone https://github.com/exelban/JAM.git
cd JAM
go build -o jam cmd/jam/main.go
./jam
```

## Configuration
The application is configured via JSON or YAML file. You can find the [example](https://github.com/exelban/JAM/blob/master/example.yaml) of the configuration file in the repository.
You can set the path to the configuration file via the `--config-path` flag (`CONFIG_PATH` env) or by default it will look for the `config.yaml` file in the current directory.

## License
[MIT License](https://github.com/exelban/JAM/blob/master/LICENSE)
