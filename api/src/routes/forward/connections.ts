import { hasPermissionByToken } from "../../libs/permissions.js";
import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens, backends } = routeOptions;

  function hasPermission(
    token: string,
    permissionList: string[],
  ): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  }

  fastify.post(
    "/api/v1/forward/connections",
    {
      schema: {
        body: {
          type: "object",
          required: ["token", "id"],

          properties: {
            token: { type: "string" },
            id: { type: "number" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error
      const body: {
        token: string;
        id: number;
      } = req.body;

      if (!(await hasPermission(body.token, ["routes.visibleConn"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const forward = await prisma.forwardRule.findUnique({
        where: {
          id: body.id,
        },
      });

      if (!forward) {
        return res.status(400).send({
          error: "Could not find forward entry",
        });
      }

      if (!backends[forward.destProviderID]) {
        return res.status(400).send({
          error: "Backend not found",
        });
      }

      return {
        success: true,
        data: backends[forward.destProviderID].getAllConnections().filter((i) => {
          return i.connectionDetails.sourceIP == forward.sourceIP && i.connectionDetails.sourcePort == forward.sourcePort && i.connectionDetails.destPort == forward.destPort;
        }),
      };
    },
  );
}
