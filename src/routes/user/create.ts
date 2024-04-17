import type { PrismaClient } from "@prisma/client";
import type { FastifyInstance } from "fastify";

import { hash } from "bcrypt";

import { ServerOptions, SessionToken } from "../../libs/types.js";
import { generateToken } from "../../libs/generateToken.js";

export function route(fastify: FastifyInstance, prisma: PrismaClient, tokens: Record<number, SessionToken[]>, options: ServerOptions) {
  // TODO: Permissions
  
  fastify.post("/api/v1/users/create", {
    schema: {
      body: {
        type: "object",
        required: ["name", "email", "password"],

        properties: {
          name: { type: "string" },
          email: { type: "string" },
          password: { type: "string" }
        }
      }
    }
  }, async(req, res) => {
    // @ts-ignore
    const body: {
      name: string,
      email: string,
      password: string
    } = req.body;

    if (!options.isSignupEnabled) {
      return res.status(400).send({
        error: "Signing up is not enabled at this time."
      });
    };

    const userSearch = await prisma.user.findFirst({
      where: {
        email: body.email
      }
    });

    if (userSearch) {
      return res.status(400).send({
        error: "User already exists"
      })
    };

    const saltedPassword: string = await hash(body.password, 15);

    let userData = {
      name: body.name,
      email: body.email,
      password: saltedPassword,
      rootToken: null
    };

    if (options.allowUnsafeGlobalTokens) {
      userData.rootToken = generateToken() as unknown as null;
    }

    const userCreateResults = await prisma.user.create({
      data: userData
    });

    // FIXME(?): Redundant checks
    if (options.allowUnsafeGlobalTokens) {
      return {
        success: true,
        token: userCreateResults.rootToken
      };
    } else {
      const generatedToken = generateToken();

      tokens[userCreateResults.id] = [];

      tokens[userCreateResults.id].push({
        createdAt: Date.now(),
        expiresAt: Date.now() + (30 * 60_000),

        token: generatedToken
      });

      return {
        success: true,
        token: generatedToken
      };
    };
  });
}