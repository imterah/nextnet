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
    "/api/v1/backends/lookup",
    {
      schema: {
        body: {
          type: "object",
          required: ["token"],

          properties: {
            token: { type: "string" },
            id: { type: "number" },
            name: { type: "string" },
            description: { type: "string" },
            backend: { type: "string" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error: Fastify routes schema parsing is trustworthy, so we can "assume" invalid types
      const body: {
        token: string;
        id?: number;
        name?: string;
        description?: string;
        backend?: string;
      } = req.body;

      if (
        !(await hasPermission(body.token, [
          "backends.visible", // wtf?
        ]))
      ) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const canSeeSecrets = await hasPermission(body.token, [
        "backends.secretVis",
      ]);

      const prismaBackends = await prisma.desinationProvider.findMany({
        where: {
          id: body.id,
          name: body.name,
          description: body.description,
          backend: body.backend,
        },
      });

      return {
        success: true,
        data: prismaBackends.map(i => ({
          id: i.id,

          name: i.name,
          description: i.description,

          backend: i.backend,
          connectionDetails: canSeeSecrets ? i.connectionDetails : "",

          logs: backends[i.id].logs,
        })),
      };
    },
  );
}
