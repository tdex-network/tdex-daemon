# tdex-daemon

[![Go Report Card](https://goreportcard.com/badge/github.com/tdex-network/tdex-daemon)](https://goreportcard.com/report/github.com/tdex-network/tdex-daemon)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tdex-network/tdex-daemon)](https://pkg.go.dev/github.com/tdex-network/tdex-daemon)
[![Release](https://img.shields.io/github/release/tdex-network/tdex-daemon.svg)](https://github.com/tdex-network/tdex-daemon/releases/latest)

Go implementation of the TDex Daemon

## 📄 Usage

In-depth documentation for installing and using the tdex-daemon is available at [docs.tdex.network](https://dev.tdex.network/docs/provider/intro)


## 🛣 Roadmap

* [x] Swap protocol
* [x] Trade protocol
* [x] Confidential support
* [x] Automated Market Making
* [x] Pluggable Market Making


## 🖥 Local Development

Below is a list of commands you will probably find useful for development.

### Requirements

* Go (^1.16.*)

### Run daemon

Builds `tdexd` as static binary and runs the project with default configuration.

```bash
$ make run
```

### Build daemon

Builds `tdexd` as static binary in the `./build` folder

```bash
$ make build
```

### Build CLI

Builds `tdex` as static binary in the `./build` folder

```bash
$ make build-cli
```

### Build unlocker

Builds `unlockerd` as static binary in the `./build` folder

```bash
$ make build-unlocker
```

### Build tdexdconnect

Builds `tdexdconnect` as a static binary in the `./build` folder

```bash
$ make build-tdexdconnect
```

### Build and Run with docker

Build and use `tdex` with docker.

#### Build tdexd docker image

_At the root of the repository_

```bash
$ docker build --pull --rm -f "Dockerfile" -t tdexd:latest "."
```

#### Run the daemon

```bash
$ docker run -d -it --name tdexd -p 9945:9945 -p 9000:9000 -v `pwd`/tdexd:/.tdex-daemon tdexd:latest
```

#### Use the CLI

```bash
$ alias tdex="docker exec -it tdexd tdex"
```

#### Use the unlocker

```bash
$ alias unlockerd="docker exec -it tdexd unlockerd"
```

### Use tdexdconnect

```bash
$ alias tdexdconnect="docker exec tdexd tdexdconnect"
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

