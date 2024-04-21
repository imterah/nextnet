import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import { ServerOptions, SessionToken } from "../../libs/types.js";
import { hasPermissionByToken } from "../../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, options: ServerOptions) {
  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };
  
  /**
   * Creates a new backend to use
   */
  fastify.post("/api/v1/users/remove", {
    schema: {
      body: {
        type: "object",
        required: ["token", "uid"],

        properties: {
          token:             { type: "string" },
          uid:               { type: "number" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string,
      uid: number
    } = req.body;

    if (!await hasPermission(body.token, [
      "users.remove"
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    await prisma.permission.deleteMany({
      where: {
        userID: body.uid
      }
    });

    await prisma.user.delete({
      where: {
        id: body.uid
      }
    });

    return {
      success: true
    }
  });
};