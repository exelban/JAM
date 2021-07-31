FROM exelban/baseimage:golang-1.16 as build-app

WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o ./bin/main ./app/main.go


FROM exelban/baseimage:alpine-latest
WORKDIR /srv
COPY --from=build-app /app/bin/main /srv/main
ENTRYPOINT ./main