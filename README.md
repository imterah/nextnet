<h1 align="center">Hermes</h1>

<p align="center">
  <img src="https://img.shields.io/badge/built-with_docker-purple" alt="Docker Badge"/>
  <img src="https://img.shields.io/badge/built-with_Go-blue" alt="Golang Badge">
  <img src="https://img.shields.io/badge/license-BSD--3--Clause-green" alt="License Badge (licensed under BSD-3-Clause)"/>
</p>

<p align="center">
  <b>Port forwarding across boundaries.</b>
</p>

<h2 align="center">Local Development</h2>

> [!NOTE]
> Using [Nix](https://builtwithnix.org) is recommended for the development environment. If you're not using it, install Go.

1. First off, clone the repository: `git clone --recurse-submodules https://git.terah.dev/imterah/hermes`

2. Then, check if you have a working Nix environment if you're using Nix.

3. Last, Run `nix-shell`, or alternatively `source init.sh` if you're not using Nix.

<h3 align="center">API Development</h3>

1. After that, run the backend build script: `./build.sh`.

2. Then, go into the `api/` directory, and then start it up: `go run . -b ../backends.dev.json`

<h2 align="center">Production Deployment</h2>

> [!WARNING]
> Deploying using [Docker Compose](https://docs.docker.com/compose/) is the only officially supported deployment method.

1. Copy and change the default password (or username & db name too) from the template file `prod-docker.env`:
  ```bash
  sed "s/POSTGRES_PASSWORD=hermes/POSTGRES_PASSWORD=$(head -c 500 /dev/random | sha512sum | cut -d " " -f 1)/g" prod-docker.env > .env
  ```

2. Build the docker stack: `docker compose --env-file .env up -d`

<h2 align="center">Troubleshooting</h2>

This has been moved [here.](docs/troubleshooting.md)

<h2 align="center">Documentation</h2>

Go to the `docs/` folder.
