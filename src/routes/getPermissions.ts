import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import { ServerOptions, SessionToken } from "../libs/types.js";
import { hasPermission, getUID } from "../libs/permissions.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, options: ServerOptions) {
  /**
   * Logs in to a user account.
   */
  fastify.post("/api/v1/getPermissions", {
    schema: {
      body: {
        type: "object",
        required: ["token"],

        properties: {
          token: { type: "string" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string
    } = req.body;

    const uid = await getUID(body.token, tokens, prisma);

    if (!await hasPermission([
      "permissions.see"
    ], uid, prisma)) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    const permissionsRaw = await prisma.permission.findMany({
      where: {
        userID: uid
      }
    });

    return {
      success: true,
      // Get the ones that we have, and transform them into just their name
      data: permissionsRaw.filter((i) => i.has).map((i) => i.permission)
    }
  });
}
