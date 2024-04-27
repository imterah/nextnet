import { compare } from "bcrypt";

import { generateToken } from "../../libs/generateToken.js";
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

    const passwordIsValid = compare(userSearch.password, body.password);
    
    if (!passwordIsValid) return res.status(403).send({
      error: "Email or password is incorrect"
    });

    const token = generateToken();
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