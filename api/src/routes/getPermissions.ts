import { hasPermission, getUID } from "../libs/permissions.js";
import type { RouteOptions } from "../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens } = routeOptions;

  /**
   * Logs in to a user account.
   */
  fastify.post(
    "/api/v1/getPermissions",
    {
      schema: {
        body: {
          type: "object",
          required: ["token"],

          properties: {
            token: { type: "string" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error: Fastify routes schema parsing is trustworthy, so we can "assume" invalid types
      const body: {
        token: string;
      } = req.body;

      const uid = await getUID(body.token, tokens, prisma);

      if (!(await hasPermission(["permissions.see"], uid, prisma))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const permissionsRaw = await prisma.permission.findMany({
        where: {
          userID: uid,
        },
      });

      return {
        success: true,
        // Get the ones that we have, and transform them into just their name
        data: permissionsRaw.filter(i => i.has).map(i => i.permission),
      };
    },
  );
}
