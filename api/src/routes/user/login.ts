import { compare } from "bcrypt";

import { generateRandomData } from "../../libs/generateRandom.js";
import type { RouteOptions } from "../../libs/types.js";

export function route(routeOptions: RouteOptions) {
  const {
    fastify,
    prisma,
    tokens
  } = routeOptions;

  /**
   * Logs in to a user account.
   */
  fastify.post("/api/v1/users/login", {
    schema: {
      body: {
        type: "object",
        required: ["email", "password"],

        properties: {
          email:    { type: "string" },
          password: { type: "string" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      email: string,
      password: string
    } = req.body;

    const userSearch = await prisma.user.findFirst({
      where: {
        email: body.email
      }
    });

    if (!userSearch) return res.status(403).send({
      error: "Email or password is incorrect"
    });

    const passwordIsValid = await compare(body.password, userSearch.password);
    
    if (!passwordIsValid) return res.status(403).send({
      error: "Email or password is incorrect"
    });

    const token = generateRandomData();
    if (!tokens[userSearch.id]) tokens[userSearch.id] = [];
    
    tokens[userSearch.id].push({
      createdAt: Date.now(),
      expiresAt: Date.now() + (30 * 60_000),

      token
    });

    return {
      success: true,
      token
    }
  });
}