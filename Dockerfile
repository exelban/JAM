FROM exelban/baseimage:golang-1.17 as build-app

WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o ./bin/main ./main.go


FROM exelban/baseimage:alpine-latest
WORKDIR /srv
COPY --from=build-app /app/bin/main /srv/main
ENTRYPOINT ./main
