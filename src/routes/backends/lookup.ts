import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import { ServerOptions, SessionToken } from "../../libs/types.js";
import { hasPermissionByToken } from "../../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, options: ServerOptions) {
  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };

  /**
   * Creates a new route to use
   */
  fastify.post("/api/v1/backends/lookup", {
    schema: {
      body: {
        type: "object",
        required: ["token"],

        properties: {
          token:             { type: "string" },
          id:                { type: "number" },
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
      id?: number,
      name?: string,
      description?: string,
      backend?: string
    } = req.body;

    if (!await hasPermission(body.token, [
      "backends.visible" // wtf?
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    const canSeeSecrets = await hasPermission(body.token, [
      "backends.secretVis"
    ]);
    
    const backends = await prisma.desinationProvider.findMany({
      where: {
        id: body.id,
        name: body.name,
        description: body.description,
        backend: body.backend
      }
    });

    return {
      success: true,
      data: backends.map((i) => ({
        name: i.name,
        description: i.description,

        backend: i.backend,
        connectionDetails: canSeeSecrets ? i.connectionDetails : ""
      }))
    }
  });
}