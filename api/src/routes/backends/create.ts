import { hasPermissionByToken } from "../../libs/permissions.js";
import type { RouteOptions } from "../../libs/types.js";

import { backendProviders } from "../../backendimpl/index.js";
import { backendInit } from "../../libs/backendInit.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens, backends } = routeOptions;

  const logWrapper = (arg: string) => fastify.log.info(arg);
  const errorWrapper = (arg: string) => fastify.log.error(arg);

  function hasPermission(
    token: string,
    permissionList: string[],
  ): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  }

  /**
   * Creates a new backend to use
   */
  fastify.post(
    "/api/v1/backends/create",
    {
      schema: {
        body: {
          type: "object",
          required: ["token", "name", "backend", "connectionDetails"],

          properties: {
            token: { type: "string" },
            name: { type: "string" },
            description: { type: "string" },
            backend: { type: "string" },
            connectionDetails: { type: "string" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error: Fastify routes schema parsing is trustworthy, so we can "assume" invalid types
      const body: {
        token: string;
        name: string;
        description?: string;
        connectionDetails: string;
        backend: string;
      } = req.body;

      if (!(await hasPermission(body.token, ["backends.add"]))) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      if (!backendProviders[body.backend]) {
        return res.status(400).send({
          error: "Unknown/unsupported/deprecated backend!",
        });
      }

      const connectionDetailsValidityCheck = backendProviders[
        body.backend
      ].checkParametersBackendInstance(body.connectionDetails);

      if (!connectionDetailsValidityCheck.success) {
        return res.status(400).send({
          error:
            connectionDetailsValidityCheck.message ??
            "Unknown error while attempting to parse connectionDetails (it's on your side)",
        });
      }

      const backend = await prisma.desinationProvider.create({
        data: {
          name: body.name,
          description: body.description,

          backend: body.backend,
          connectionDetails: body.connectionDetails,
        },
      });

      const init = await backendInit(
        backend,
        backends,
        prisma,
        logWrapper,
        errorWrapper,
      );

      if (!init) {
        // TODO: better error code
        return res.status(504).send({
          error: "Backend is created, but failed to initalize correctly",
          id: backend.id,
        });
      }

      return {
        success: true,
        id: backend.id,
      };
    },
  );
}
