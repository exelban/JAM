FROM exelban/baseimage:golang-latest as build-app

WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o ./bin/main ./main.go

FROM exelban/baseimage:alpine-latest
EXPOSE 8822
WORKDIR /app
COPY --from=build-app /app/bin/main /app/main
ENTRYPOINT ./main