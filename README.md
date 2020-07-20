# Birrometro bot

This bot keeps track of beer debts in a group.

**Work in progress**: see "Limitations and TODOs"

## Build requirements

Tested with Go 1.14 on Debian GNU/Linux. It should work in all major platforms where Go lives as there are no native libraries.

## Usage

Build the bot using `make bot`. Then issue `./bot --help` to see which environment variables are supported.
Bot Token and the database path are required.

## Docker

There is a [ready-to-go Docker image in Docker Hub](https://hub.docker.com/r/sysadminita/birrometro_bot).

## Kubernetes

There are some examples of Kubernetes configuration in `k8s/`. Secrets like bot-token should be created in
`birrometro-bot-secrets` secret.

# Limitations and TODOs

* [ ] Make the bot multilanguage
* [ ] Add support multiple groups
* [ ] Add support for administrative commands
* [ ] Add support for choco bar
