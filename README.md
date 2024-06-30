<h1 align="center">NextNet</h1>

<p align="center">
   <a href="https://builtwithnix.org"><img src="https://builtwithnix.org/badge.svg" alt="built with nix" height="20"/></a>
   <img src="https://img.shields.io/github/license/greysoh/nextnet" alt="License Badge"/>
</p>

<br>

**NextNet is a dashboard to manage portforwarding technologies.**

<h2 align="center">Local Development</h2>

> [!NOTE]
> Using [nix](https://builtwithnix.org) is recommended. If you're not using Nix, install PostgreSQL, Node.JS, and `lsof`.

1. First, check if you have a working Nix environment if you're using Nix.

2. Run `nix-shell`, or alternatively `source init.sh` if you're not using Nix.

<h3 align="center">API Development</h3>

1. After that, run the project in development mode: `npm run dev`.

2. If you want to explore your database, run `npx prisma studio` to open the database editor.

<h2 align="center">Production Deployment</h2>

> [!WARNING]  
> Deploying using docker compose is the only officially supported deployment method. Here be dragons!

1. Copy and change the default password (or username & db name too) from the template file `prod-docker.env`:
   ```bash
   sed "s/POSTGRES_PASSWORD=nextnet/POSTGRES_PASSWORD=$(head -c 500 /dev/random | sha512sum | cut -d " " -f 1)/g" prod-docker.env > .env
   ```
  
2. Build the docker stack: `docker compose --env-file .env up -d`

<h2 align="center">Troubleshooting</h2>

* I'm using SSH-based tunneling, and I can't reach any of the tunnels publicly, but can reach them on the SSH server.

  - Be sure to enable GatewayPorts in your sshd config (in `/etc/ssh/sshd_config` on most systems). Also, be sure to check your firewall rules on your system and your network.