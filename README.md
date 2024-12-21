<h1 align="center">Hermes</h1>

<p align="center">
  <img src="https://img.shields.io/badge/built-with_docker-purple" alt="Docker Badge"/>
  <img src="https://img.shields.io/badge/built-with_Go-blue" alt="Golang Badge">
  <img src="https://img.shields.io/badge/license-BSD--3--Clause-green" alt="License Badge"/>
</p>

<br>

<p align="center">
  <b>Port forwarding across boundaries.</b>
</p>

<h2 align="center">Local Development</h2>

> [!NOTE]
> Using [Nix](https://builtwithnix.org) is recommended for the development environment. If you're not using it, install Go and NodeJS.
  Using [Docker](https://www.docker.com/) is required for database configuration.

1. First, make sure you have a sane copy of Docker installed, and make sure the copy of Docker works.

2. Secondly, check if you have a working Nix environment if you're using Nix.

3. Lastly, Run `nix-shell`, or alternatively `source init.sh` if you're not using Nix.

<h3 align="center">API Development</h3>

1. After that, run the project in development mode: `npm run dev`.

2. If you want to explore your database, run `npx prisma studio` to open the database editor.

<h2 align="center">Production Deployment</h2>

> [!WARNING]
> Deploying using [Docker Compose](https://docs.docker.com/compose/) is the only officially supported deployment method.

1. Copy and change the default password (or username & db name too) from the template file `prod-docker.env`:
  ```bash
  sed "s/POSTGRES_PASSWORD=nextnet/POSTGRES_PASSWORD=$(head -c 500 /dev/random | sha512sum | cut -d " " -f 1)/g" prod-docker.env > .env
  ```

2. Build the docker stack: `docker compose --env-file .env up -d`

<h2 align="center">Troubleshooting</h2>

* I'm using SSH tunneling, and I can't reach any of the tunnels publicly.

  - Be sure to enable GatewayPorts in your sshd config (in `/etc/ssh/sshd_config` on most systems). Also, be sure to check your firewall rules on your system and your network.
