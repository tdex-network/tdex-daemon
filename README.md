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

## Requirements

* Go (^1.15.2)

### Run

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

### Test

```bash
# Short testing
$ make test

# integration testing
$ make integrationtest
```