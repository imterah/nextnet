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

  /**
   * Creates a new route to use
   */
  fastify.post(
    "/api/v1/forward/stop",
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

      if (!(await hasPermission(body.token, ["routes.stop"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const forward = await prisma.forwardRule.findUnique({
        where: {
          id: body.id,
        },
      });

      if (!forward)
        return res.status(400).send({
          error: "Could not find forward entry",
        });

      if (!backends[forward.destProviderID])
        return res.status(400).send({
          error: "Backend not found",
        });

      // Other restrictions in place make it so that it MUST be either TCP or UDP
      // @ts-expect-error
      const protocol: "tcp" | "udp" = forward.protocol;

      backends[forward.destProviderID].removeConnection(
        forward.sourceIP,
        forward.sourcePort,
        forward.destPort,
        protocol,
      );

      return {
        success: true,
      };
    },
  );
}
