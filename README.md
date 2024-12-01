<h1 align="center">NextNet</h1>

<p align="center">
   <a href="https://builtwithnix.org"><img src="https://builtwithnix.org/badge.svg" alt="built with nix" height="20"/></a>
   <img src="https://img.shields.io/github/license/greysoh/nextnet" alt="License Badge"/>
</p>

<br>

**NextNet is a dashboard to manage portforwarding technologies.**

<h2 align="center">⚠️ Deprecation Warning ⚠️</h2>

NextNet in its current state is going to be deprecated and slowly rewritten into Go, with a more modular approach that's
similar to Tor's modular pluggable transports system.

**What will change for end users?** There will be an export feature added to the legacy codebase, and an import feature 
for the new codebase. You will need to upgrade to intermediate versions that allow for this. After this one-time process
is done, you won't have to run it ever again.

This will also lead for performance benefits (hopefully). The flagship backend for now is SSH. The implementation for that
is node-ssh, which... isn't the fastest thing ever, since it reimplements the SSH protocol in pure JS.

**What will change for developers?** The LOM and API will be rewritten in Go slowly. See issue [#1](https://git.greysoh.dev/imterah/nextnet/issues/1) on my Git server as a tracking issue for this. Except for new Go code and migration code, this project is
on a feature freeze effective *immediately*.

The code for this is on branch `dev`, like usual. However, this warning and all the old code is on the `legacy` branch,
which is currently the default branch. If you're a developer, be careful and make sure you're commiting to the right branch,
*especially* if you've just cloned the code.

Additionally, the Git server has moved from `https://github.com/imterah/nextnet.git` to 
`https://git.greysoh.dev/imterah/nextnet.git`. Be sure to update your remotes. All PRs and issues on GitHub *will* be ignored.

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

* I'm using the SSH tunneling, and I can't reach any of the tunnels publicly.

  - Be sure to enable GatewayPorts in your sshd config (in `/etc/ssh/sshd_config` on most systems). Also, be sure to check your firewall rules on your system and your network.