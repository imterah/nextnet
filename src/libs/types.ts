import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import type { BackendBaseClass } from "../backendimpl/base.js";

export type ServerOptions = {
  isSignupEnabled: boolean;
  isSignupAsAdminEnabled: boolean;

  allowUnsafeGlobalTokens: boolean;
}

// NOTE: Someone should probably use Redis for this, but this is fine...
export type SessionToken = {
  createdAt: number,
  expiresAt: number, // Should be (createdAt + (30 minutes))
  
  token: string
};

export type RouteOptions = {
  fastify: FastifyInstance,
  prisma: PrismaClient,
  tokens: Record<number, SessionToken[]>,
  
  options: ServerOptions,
  backends: Record<number, BackendBaseClass>
};