FROM exelban/baseimage:golang-latest AS build-app

ARG VERSION

WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN if [ -z "$VERSION" ]; then  \
    VERSION="$(/script/build_time.sh)"; \
    fi && \
    go build -ldflags "-X main.version=$VERSION" -o bin/main

FROM exelban/baseimage:alpine-latest
EXPOSE 8822
WORKDIR /app
COPY --from=build-app /app/bin/main /app/main
ENTRYPOINT ["./main"]