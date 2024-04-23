import { hasPermissionByToken } from "../../libs/permissions.js";
import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const {
    fastify,
    prisma,
    tokens
  } = routeOptions;

  function hasPermission(token: string, permissionList: string[]): Promise<boolean> {
    return hasPermissionByToken(permissionList, token, tokens, prisma);
  };

  /**
   * Creates a new route to use
   */
  fastify.post("/api/v1/forward/create", {
    schema: {
      body: {
        type: "object",
        required: ["token", "name", "sourceIP", "sourcePort", "destinationPort", "providerID"],

        properties: {
          token: { type: "string" },

          name: { type: "string" },
          description: { type: "string" },

          sourceIP: { type: "string" },
          sourcePort: { type: "number" },

          destinationPort: { type: "number" },

          providerID: { type: "number" },
          autoStart: { type: "boolean" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      token: string,

      name: string,
      description?: string,

      sourceIP: string,
      sourcePort: number,

      destinationPort: number,

      providerID: number,

      autoStart?: boolean
    } = req.body;

    if (!await hasPermission(body.token, [
      "routes.add"
    ])) {
      return res.status(403).send({
        error: "Unauthorized"
      });
    };

    const lookupIDForDestProvider = await prisma.desinationProvider.findUnique({
      where: {
        id: body.providerID
      }
    });

    if (!lookupIDForDestProvider) return res.status(400).send({
      error: "Could not find provider"
    });

    await prisma.forwardRule.create({
      data: {
        name: body.name,
        description: body.description,

        sourceIP: body.sourceIP,
        sourcePort: body.sourcePort,

        destPort: body.destinationPort,

        destProviderID: body.providerID,

        enabled: Boolean(body.autoStart)
      }
    });

    return {
      success: true
    }
  });
}