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
3. After that, run the project in development mode: `npm run dev`.
4. If you want to explore your database, run `npx prisma studio` to open the database editor.