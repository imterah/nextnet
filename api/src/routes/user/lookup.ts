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

  fastify.post(
    "/api/v1/users/lookup",
    {
      schema: {
        body: {
          type: "object",
          required: ["token"],

          properties: {
            token: { type: "string" },
            id: { type: "number" },
            name: { type: "string" },
            email: { type: "string" },
            username: { type: "string" },
            isServiceAccount: { type: "boolean" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-ignore
      const body: {
        token: string;
        id?: number;
        name?: string;
        email?: string;
        username?: string;
        isServiceAccount?: boolean;
      } = req.body;

      if (!(await hasPermission(body.token, ["users.lookup"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const users = await prisma.user.findMany({
        where: {
          id: body.id,
          name: body.name,
          email: body.email,
          username: body.username,
          isRootServiceAccount: body.isServiceAccount,
        },
      });

      return {
        success: true,
        data: users.map(i => ({
          id: i.id,
          name: i.name,
          email: i.email,
          isServiceAccount: i.isRootServiceAccount,
          username: i.username
        })),
      };
    },
  );
}
