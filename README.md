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

Below is a list of commands you will likely find useful for development.

### Requirements

* [Golang](https://go.dev/) (^1.16.*)
* [Ocean wallet](https://github.com/vulpemventures/ocean)

### Run daemon (dev mode)

[Start](https://github.com/vulpemventures/ocean/#local-run) the ocean wallet.

Start the daemon:

```bash
$ make run
```

### Build daemon

Build `tdexd` as a static binary in the `./build` folder

```bash
$ make build
```

### Build CLI

Build `tdex` as a static binary in the `./build` folder

```bash
$ make build-cli
```

### Build and Run with docker

Start oceand and tdexd services as docker contaniner.

#### Start oceand and tdexd

Start `oceand` and `tdexd` containters:

```bash
$ docker-compose -f resources/compose/docker-compose.yml up -d oceand tdexd
```

#### Use the CLI

```bash
$ alias tdex="docker exec tdexd tdex"

# Configure the CLI
$ tdex config init --no-tls --no-macaroons

# Use the CLI
$ tdex status
$ tdex --help
```

### Test

```bash
# Unit testing
$ make test

# Integration testing
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

