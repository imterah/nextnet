import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import { ServerOptions, SessionToken } from "../../libs/types.js";
import { hasPermissionByToken } from "../../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, options: ServerOptions) {
  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };

  fastify.post("/api/v1/users/lookup", {
    schema: {
      body: {
        type: "object",
        required: ["token"],

        properties: {
          token:            { type: "string" },
          name:             { type: "string" },
          email:            { type: "string" },
          isServiceAccount: { type: "boolean" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string,
      name?: string,
      email?: string,
      isServiceAccount?: boolean
    } = req.body;

    if (!await hasPermission(body.token, [
      "users.lookup"
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    const users = await prisma.user.findMany({
      where: {
        name: body.name,
        email: body.email,
        isRootServiceAccount: body.isServiceAccount
      }
    });

    return {
      success: true,
      data: users.map((i) => ({
        name: i.name,
        email: i.email,
        isServiceAccount: i.isRootServiceAccount
      }))
    }
  });
}