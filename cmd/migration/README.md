# Migration

The migration service translates a daemon's datadir with a specific major version to the very next one.

Currently supported versions:
* from `v0` to `v1`, requires a daemon with exact verion `v0.9.1`

## Build

```sh
$ make build-migration
$ alias migration=./build/migration-<os>-<arch>
```

## Usage

* Standalone binary version
  ```sh
  $ migration --password <wallet_password> [--source-version, --v0-datadir, --v1-datadir, --ocean-datadir, --no-backup]
  ```
* Docker version
  ```sh
  # Pull the latest image
  $ docker pull ghcr.io/tdex-network/tdexd:latest
  $ docker run --rm --volume path/to/datadir:/home/tdex/.tdex-daemon --entrypoint 'tdex-migration' ghcr.io/tdex-network/tdexd:latest --password <wallet_password> --/home/tdex/.tdex-daemon/oceand
  ```

You can get more info about the usage of the flags by running `tdex-migration --help` at anytime.

Once the migration is completed, it is enough to start up [ocean](https://github.com/vulpemventures/ocean) and configure it to use the newly created datadir and using the filesystem-based DB (badger).  

NOTE: currently only migration from v0 to v1 is supported, but in case more are added in the future (like for example migration from v1 to v2), you can select the version to be migrated by using the flag `--source-version`. You don't have to specify a dest version because this tool allow migrating only from a major version to the very next one.