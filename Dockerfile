FROM exelban/baseimage:node-14 as build-web

WORKDIR /app/

COPY admin/package*.json ./
COPY admin/yarn.lock ./
RUN yarn --silent

COPY admin/ .
RUN yarn build

FROM exelban/baseimage:golang-1.17 as build-app

WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY --from=build-web /app/dist /app/admin/dist
COPY . .

RUN go build -o ./bin/main ./main.go

FROM exelban/baseimage:alpine-latest
WORKDIR /srv
COPY --from=build-app /app/bin/main /srv/main
ENTRYPOINT ./main