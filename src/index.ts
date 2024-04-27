import process from "node:process";

import { PrismaClient } from '@prisma/client';
import Fastify from "fastify";

import type { ServerOptions, SessionToken, RouteOptions } from "./libs/types.js";
import type { BackendBaseClass } from "./backendimpl/base.js";``

import { route as getPermissions } from "./routes/getPermissions.js";

import { route as backendCreate } from "./routes/backends/create.js";
import { route as backendRemove } from "./routes/backends/remove.js";
import { route as backendLookup } from "./routes/backends/lookup.js";

import { route as forwardCreate } from "./routes/forward/create.js";
import { route as forwardRemove } from "./routes/forward/remove.js";
import { route as forwardLookup } from "./routes/forward/lookup.js";

import { route as userCreate } from "./routes/user/create.js";
import { route as userRemove } from "./routes/user/remove.js";
import { route as userLookup } from "./routes/user/lookup.js";
import { route as userLogin } from "./routes/user/login.js";
import { backendInit } from "./libs/backendInit.js";

const prisma = new PrismaClient();

const isSignupEnabled: boolean   = Boolean(process.env.IS_SIGNUP_ENABLED);
const unsafeAdminSignup: boolean = Boolean(process.env.UNSAFE_ADMIN_SIGNUP);

const noUsersCheck: boolean = await prisma.user.count() == 0;

if (unsafeAdminSignup) {
  console.error("WARNING: You have admin sign up on! This means that anyone that signs up will have admin rights!");
}

const serverOptions: ServerOptions = {
  isSignupEnabled: isSignupEnabled ? true : noUsersCheck,
  isSignupAsAdminEnabled: unsafeAdminSignup ? true : noUsersCheck,

  allowUnsafeGlobalTokens: process.env.NODE_ENV != "production"
};

const sessionTokens: Record<number, SessionToken[]> = {};
const backends: Record<number, BackendBaseClass> = {};

const fastify = Fastify({
  logger: true
});

const routeOptions: RouteOptions = {
  fastify: fastify,
  prisma: prisma,
  tokens: sessionTokens,
  options: serverOptions,

  backends: backends
};

console.log("Initializing forwarding rules...");

const createdBackends = await prisma.desinationProvider.findMany();

for (const backend of createdBackends) {
  console.log(`Running init steps for ID '${backend.id}' (${backend.name})`);
  const init = await backendInit(backend, backends, prisma);  

  if (init) console.log("Init successful.");
}

console.log("Done.");

getPermissions(routeOptions);
 
backendCreate(routeOptions);
backendRemove(routeOptions);
backendLookup(routeOptions);

forwardCreate(routeOptions);
forwardRemove(routeOptions);
forwardLookup(routeOptions);

userCreate(routeOptions);
userRemove(routeOptions);
userLookup(routeOptions);
userLogin(routeOptions);

// Run the server!
try {
  await fastify.listen({ port: 3000 });
} catch (err) {
  fastify.log.error(err);
  process.exit(1);
}