# tdex-daemon

[![Go Report Card](https://goreportcard.com/badge/github.com/tdex-network/tdex-daemon?style=flat-square)](https://goreportcard.com/report/github.com/tdex-network/tdex-daemon)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tdex-network/tdex-daemon)](https://pkg.go.dev/github.com/tdex-network/tdex-daemon)
[![Release](https://img.shields.io/github/release/tdex-network/tdex-daemon.svg?style=flat-square)](https://github.com/tdex-network/tdex-daemon/releases/latest)

Go implementation of the TDex Daemon

## 📄 Usage

In-depth documentation for installing and using the tdex-daemon is available at [docs.tdex.network](https://docs.tdex.network/tdex-daemon.html)


## 🛣 Roadmap

* [x] Swap protocol
* [x] Trade protocol
* [x] Confidential support
* [x] Automated Market Making
* [x] Pluggable Market Making


## 🖥 Local Development

Below is a list of commands you will probably find useful for development.

### Requirements

* Go (^1.15.*)

### Run daemon

Builds `tdexd` as static binary and runs the project with default configuration.

```bash
# Max OSX
$ make run-mac

# Linux
$ make run-linux
```

### Build daemon

Builds `tdexd` as static binary in the `./build` folder

```bash
# Max OSX
$ make build-mac

# Linux
$ make build-linux

# ARM
$ make build-arm
```

### Build CLI

Builds `tdex` as static binary in the `./build` folder

```bash
# Max OSX
$ make build-cli-mac

# Linux
$ make build-cli-linux

# ARM
$ make build-cli-arm
```

### Build and Run with docker

Build and use `tdex` with docker.

#### Build tdexd docker image

_At the root of the repository_

```bash
docker build --pull --rm -f "Dockerfile" -t tdexd:latest "."
```

#### Run the daemon

The following command launch the daemon targetting a regtest network hosted on your computer.

```bash
docker run -it --name tdexd \
    -p 9945:9945 -p 9000:9000 \
    -v `pwd`/tdexd:/.tdex-daemon \
    -e TDEX_NETWORK=regtest \
    -e TDEX_EXPLORER_ENDPOINT=http://127.0.0.1:3001 \
    -e TDEX_FEE_ACCOUNT_BALANCE_TRESHOLD=1000 \
    -e TDEX_BASE_ASSET=5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225 \
    -e TDEX_LOG_LEVEL=5 \
    tdexd:latest
```

#### Use the CLI

```bash
alias tdex-cli="docker exec -it tdex tdex"
```

### Test

```bash
# Short testing
$ make test

# integration testing
$ make integrationtest
```

## Release

Precompiled binaries are published with each [release](https://github.com/tdex-network/tdex-daemon/releases).

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the
[tags on this repository](https://github.com/tdex-network/tdex-daemon/tags). 

## License

This project is licensed under the MIT License - see the
[LICENSE](https://github.com/tdex-network/tdex-daemon/blob/master/LICENSE) file for details.

