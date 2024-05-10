import { compare } from "bcrypt";

import { generateRandomData } from "../../libs/generateRandom.js";
import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const { fastify, prisma, tokens } = routeOptions;

  /**
   * Logs in to a user account.
   */
  fastify.post(
    "/api/v1/users/login",
    {
      schema: {
        body: {
          type: "object",
          required: ["password"],

          properties: {
            email: { type: "string" },
            username: { type: "string" },
            password: { type: "string" }
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error
      const body: {
        email?: string;
        username?: string;
        password: string;
      } = req.body;

      if (!body.email && !body.username) return res.status(400).send({
        error: "missing both email and username. please supply at least one."
      });

      const userSearch = await prisma.user.findFirst({
        where: {
          email: body.email,
          username: body.username
        },
      });

      if (!userSearch)
        return res.status(403).send({
          error: "Email or password is incorrect",
        });

      const passwordIsValid = await compare(body.password, userSearch.password);

      if (!passwordIsValid)
        return res.status(403).send({
          error: "Email or password is incorrect",
        });

      const token = generateRandomData();
      if (!tokens[userSearch.id]) tokens[userSearch.id] = [];

      tokens[userSearch.id].push({
        createdAt: Date.now(),
        expiresAt: Date.now() + 30 * 60_000,

        token,
      });

      return {
        success: true,
        token,
      };
    },
  );
}
