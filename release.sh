#!/bin/sh

set -eu

version="v0.0.3"
dev=false
dockerOnly=false

while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--version)
      version="$2"
      shift
      shift
      ;;
    -dev)
      dev=true
      version="dev"
      shift
      ;;
    -d|--docker)
      dockerOnly=true
      shift
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
  esac
done

build_darwin() {
  GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_darwin_x86_64.tar.gz -C bin jam && rm bin/jam
  GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_darwin_arm64.tar.gz -C bin jam && rm bin/jam
}
build_linux() {
  GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_linux_x86_64.tar.gz -C bin jam && rm bin/jam
  GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_linux_arm64.tar.gz -C bin jam && rm bin/jam
}
build_windows() {
  GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_windows_x86_64.tar.gz -C bin jam && rm bin/jam
  GOOS=windows GOARCH=arm64 go build -ldflags "-X main.version=$version" -o bin/jam && tar -czf release/jam_"$version"_windows_arm64.tar.gz -C bin jam && rm bin/jam
}

build_docker_hub() {
  if [ "$dev" = false ]; then
  docker buildx build --build-arg VERSION="$version" --push --platform=linux/amd64,linux/arm64 --tag=exelban/jam:latest .
  fi
  docker buildx build --build-arg VERSION="$version" --push --platform=linux/amd64,linux/arm64 --tag=exelban/jam:"$version" .
}
build_github_registry() {
  if [ "$dev" = false ]; then
  docker buildx build --build-arg VERSION="$version" --push --platform=linux/amd64,linux/arm64 --tag=ghcr.io/exelban/jam:latest .
  fi
  docker buildx build --build-arg VERSION="$version" --push --platform=linux/amd64,linux/arm64 --tag=ghcr.io/exelban/jam:"$version" .
}

if [ "$dev" = true ]; then
  printf "\033[32;1m%s\033[0m\n" "Building dev version..."
else
  printf "\033[32;1m%s\033[0m\n" "Building version ${version}..."
fi

build_docker_hub
build_github_registry

if [ "$dockerOnly" = false ]; then
  printf "\033[32;1m%s\033[0m\n" "Building precompiled binaries with version ${version}..."

  rm -rf "bin" && rm -rf "release"
  mkdir -p "release"

  echo "Building darwin..."
  build_darwin
  echo "Building linux..."
  build_linux
  echo "Building windows..."
  build_windows
  rm -rf "bin"

  printf "\033[32;1m%s\033[0m\n" "JAM ${version} was successfully build."
  open release
fi

exit 0