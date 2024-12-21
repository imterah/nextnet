import { hash } from "bcrypt";

import { permissionListEnabled } from "../../libs/permissions.js";
import { generateRandomData } from "../../libs/generateRandom.js";

import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens, options } = routeOptions;

  /**
   * Creates a new user account to use, only if it is enabled.
   */
  fastify.post(
    "/api/v1/users/create",
    {
      schema: {
        body: {
          type: "object",
          required: ["name", "email", "username", "password"],

          properties: {
            name: { type: "string" },
            username: { type: "string" },
            email: { type: "string" },
            password: { type: "string" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error: Fastify routes schema parsing is trustworthy, so we can "assume" invalid types
      const body: {
        name: string;
        email: string;
        password: string;
        username: string;
      } = req.body;

      if (!options.isSignupEnabled) {
        return res.status(403).send({
          error: "Signing up is not enabled at this time.",
        });
      }

      const userSearch = await prisma.user.findFirst({
        where: {
          email: body.email,
        },
      });

      if (userSearch) {
        return res.status(400).send({
          error: "User already exists",
        });
      }

      const saltedPassword: string = await hash(body.password, 15);

      const userData = {
        name: body.name,
        email: body.email,
        password: saltedPassword,

        username: body.username,

        permissions: {
          create: [] as {
            permission: string;
            has: boolean;
          }[],
        },
      };

      // TODO: There's probably a faster way to pull this off, but I'm lazy
      for (const permissionKey of Object.keys(permissionListEnabled)) {
        if (
          options.isSignupAsAdminEnabled ||
          permissionKey.startsWith("routes") ||
          permissionKey == "permissions.see"
        ) {
          userData.permissions.create.push({
            permission: permissionKey,
            has: permissionListEnabled[permissionKey],
          });
        }
      }

      if (options.allowUnsafeGlobalTokens) {
        // @ts-expect-error: Setting this correctly is a goddamn mess, but this is safe to an extent. It won't crash at least
        userData.rootToken = generateRandomData();
        // @ts-expect-error: Read above.
        userData.isRootServiceAccount = true;
      }

      const userCreateResults = await prisma.user.create({
        data: userData,
      });

      // FIXME(?): Redundant checks
      if (options.allowUnsafeGlobalTokens) {
        return {
          success: true,
          token: userCreateResults.rootToken,
        };
      } else {
        const generatedToken = generateRandomData();

        tokens[userCreateResults.id] = [];

        tokens[userCreateResults.id].push({
          createdAt: Date.now(),
          expiresAt: Date.now() + 30 * 60_000,

          token: generatedToken,
        });

        return {
          success: true,
          token: generatedToken,
        };
      }
    },
  );
}
