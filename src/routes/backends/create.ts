import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import type { BackendInterface } from "../../backendimpl/base.js";
import { ServerOptions, SessionToken } from "../../libs/types.js";
import { hasPermissionByToken } from "../../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, _options: ServerOptions, backends: Record<number, BackendInterface>) {
  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };
  
  /**
   * Creates a new backend to use
   */
  fastify.post("/api/v1/backends/create", {
    schema: {
      body: {
        type: "object",
        required: ["token", "name", "backend", "connectionDetails"],

        properties: {
          token:             { type: "string" },
          name:              { type: "string" },
          description:       { type: "string" },
          backend:           { type: "string" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string,
      name: string,
      description?: string,
      connectionDetails: any,
      backend: string
    } = req.body;

    if (!await hasPermission(body.token, [
      "backends.add"
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    await prisma.desinationProvider.create({
      data: {
        name: body.name,
        description: body.description,

        backend: body.backend,
        connectionDetails: JSON.stringify(body.connectionDetails)
      }
    });

    return {
      success: true
    };
  });
}