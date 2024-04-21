import process from "node:process";

import { PrismaClient } from '@prisma/client';
import Fastify from "fastify";

import { ServerOptions, SessionToken } from "./libs/types.js";

import { route as backendCreate } from "./routes/backends/create.js";

import { route as forwardCreate } from "./routes/forward/create.js";

import { route as userCreate } from "./routes/user/create.js";
import { route as userLogin } from "./routes/user/login.js";

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

const fastify = Fastify({
  logger: true
});

backendCreate(fastify, prisma, sessionTokens, serverOptions);

forwardCreate(fastify, prisma, sessionTokens, serverOptions);

userCreate(fastify, prisma, sessionTokens, serverOptions);
userLogin(fastify, prisma, sessionTokens, serverOptions);

// Run the server!
try {
  await fastify.listen({ port: 3000 });
} catch (err) {
  fastify.log.error(err);
  process.exit(1);
}