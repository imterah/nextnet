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
    "/api/v1/backends/remove",
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
      // @ts-expect-error: Fastify routes schema parsing is trustworthy, so we can "assume" invalid types
      const body: {
        token: string;
        id: number;
      } = req.body;

      if (!(await hasPermission(body.token, ["backends.remove"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      if (!backends[body.id]) {
        return res.status(400).send({
          error: "Backend not found",
        });
      }

      // Unload the backend
      if (!(await backends[body.id].stop())) {
        return res.status(400).send({
          error: "Failed to stop backend! Please report this issue.",
        });
      }

      delete backends[body.id];

      await prisma.desinationProvider.delete({
        where: {
          id: body.id,
        },
      });

      return {
        success: true,
      };
    },
  );
}
