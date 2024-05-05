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
    "/api/v1/forward/lookup",
    {
      schema: {
        body: {
          type: "object",
          required: ["token"],

          properties: {
            token: { type: "string" },
            id: { type: "number" },

            name: { type: "string" },
            protocol: { type: "string" },
            description: { type: "string" },

            sourceIP: { type: "string" },
            sourcePort: { type: "number" },
            destPort: { type: "number" },

            providerID: { type: "number" },
            autoStart: { type: "boolean" },
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
        description?: string;

        protocol?: "tcp" | "udp";

        sourceIP?: string;
        sourcePort?: number;

        destinationPort?: number;

        providerID?: number;
        autoStart?: boolean;
      } = req.body;

      if (body.protocol && body.protocol != "tcp" && body.protocol != "udp") {
        return res.status(400).send({
          error: "Protocol specified in body must be either 'tcp' or 'udp'",
        });
      }

      if (
        !(await hasPermission(body.token, [
          "routes.visible", // wtf?
        ]))
      ) {
        return res.status(403).send({
          error: "Unauthorized",
        });
      }

      const forwardRules = await prisma.forwardRule.findMany({
        where: {
          id: body.id,
          name: body.name,
          description: body.description,

          sourceIP: body.sourceIP,
          sourcePort: body.sourcePort,

          destPort: body.destinationPort,

          destProviderID: body.providerID,
          enabled: body.autoStart,
        },
      });

      return {
        success: true,
        data: forwardRules.map(i => ({
          id: i.id,
          name: i.name,
          description: i.description,

          sourceIP: i.sourceIP,
          sourcePort: i.sourcePort,

          destPort: i.destPort,

          providerID: i.destProviderID,
          autoStart: i.enabled, // TODO: Add enabled flag in here to see if we're running or not
        })),
      };
    },
  );
}
