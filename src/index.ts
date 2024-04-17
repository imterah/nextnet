import process from "node:process";

import { PrismaClient } from '@prisma/client';
import Fastify from "fastify";

import { ServerOptions, SessionToken } from "./libs/types.js";
import { route as create } from "./routes/user/create.js";

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

create(fastify, prisma, sessionTokens, serverOptions);

// Run the server!
try {
  await fastify.listen({ port: 3000 });
} catch (err) {
  fastify.log.error(err);
  process.exit(1);
}