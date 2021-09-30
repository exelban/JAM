# cheks

[![Cheks](https://serhiy.s3.eu-central-1.amazonaws.com/Github_repo/cheks/preview.png)](https://github.com/exelban/cheks/releases)
Simple monitoring for APIs and servers with dashboard and alerts.

## Install
To run the Cheks you need to have a docker and configuration file.

Cheks provides small prebuild images for different architectures and operating systems.
You could use an image from Docker Hub `exelban/cheks` or GitHub package registry `ghcr.io/exelban/cheks`.

The easiest way to run Checks is to use docker-compose:

```yaml
version: "3"

services:
  cheks:
    image: exelban/cheks
    ports:
      - "8080:8080"
    volumes:
      - ${PWD}/config.yaml:/srv/config.yaml
```

or with docker:

```shell
docker run -p 8080:8080 -v $(pwd)/config.yaml:./srv/config.yaml exelban/cheks
```

## Parameters

| Command line | Environment | Default | Description |
| ------------ | ----------- | ------- | ----------- |
| config | CONFIG | ./config.yaml | Path to the configuration file |
| auth | AUTH | false | Secure dashboard with credentials |
| username | USERNAME | | Username for the dashboard (only if auth is true). Required if AUTH=true |
| password | PASSWORD | | Password for the dashboard. If empty, will be generated on the first run |

## Configuration file
The simplest configuration file could only have a list of hosts:

```yaml
hosts:
  - url: https://github.com
  - url: https://google.com
  - url: https://facebook.com
```

Optional configuration for host:  
`name: string` - name  
`tags: [string]` - list of tags  

`retry: string` - retry interval for request. Allowed golang style durations: 10s, 60s, 3m.  
`timeout: string` - retry interval for request. Allowed golang style durations: 10s, 60s, 3m.  
`initialDelay: string` - timeout before the request will be canceled. Allowed golang style durations: 10s, 60s, 3m.  
`successThreshold: int` - number of success requests before host will be marked as live  
`failureThreshold: int` - number of failed requests before host will be marked as dead  

`success` - allows to define success request parameters as response code and response body:

```yaml
success:
  code: [200, 201, 202]
  body: {"ok": true}
```

`headers: map[string]string` - you could specify the headers which will be sent with the request

## License
[MIT License](https://github.com/exelban/cheks/blob/master/LICENSE)