import { hasPermissionByToken } from "../../libs/permissions.js";
import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens } = routeOptions;

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
    "/api/v1/forward/remove",
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
      // @ts-ignore
      const body: {
        token: string;
        id: number;
      } = req.body;

      if (!(await hasPermission(body.token, ["routes.remove"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      await prisma.forwardRule.delete({
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
