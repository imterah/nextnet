import { generateRandomData } from "../../libs/generateRandom.js";
import type { PassyFireBackendProvider } from "./index.js";

export function route(instance: PassyFireBackendProvider) {
  const { fastify } = instance;

  const proxiedPort: number = instance.options.publicPort ?? 443;

  const unsupportedSpoofedRoutes: string[] = [
    "/api/v1/tunnels/add",
    "/api/v1/tunnels/edit",
    "/api/v1/tunnels/remove",

    // TODO (greysoh): Should we implement these? We have these for internal reasons. We could expose these /shrug
    "/api/v1/tunnels/start",
    "/api/v1/tunnels/stop",

    // Same scenario for this API.
    "/api/v1/users",
    "/api/v1/users/add",
    "/api/v1/users/remove",
    "/api/v1/users/enable",
    "/api/v1/users/disable",
  ];

  fastify.get("/api/v1/static/getScopes", () => {
    return {
      success: true,
      data: {
        users: {
          add: true,
          remove: true,
          get: true,
          getPasswords: true,
        },
        routes: {
          add: true,
          remove: true,
          start: true,
          stop: true,
          get: true,
          getPasswords: true,
        },
      },
    };
  });

  for (const spoofedRoute of unsupportedSpoofedRoutes) {
    fastify.post(spoofedRoute, (req, res) => {
      return res.status(403).send({
        error: "Invalid scope(s)",
      });
    });
  }

  fastify.post(
    "/api/v1/users/login",
    {
      schema: {
        body: {
          type: "object",
          required: ["username", "password"],

          properties: {
            username: { type: "string" },
            password: { type: "string" },
          },
        },
      },
    },
    (req, res) => {
      // @ts-expect-error
      const body: {
        username: string;
        password: string;
      } = req.body;

      if (
        !instance.options.users.find(
          i => i.username == body.username && i.password == body.password,
        )
      ) {
        return res.status(403).send({
          error: "Invalid username/password.",
        });
      }

      const token = generateRandomData();

      instance.users.push({
        username: body.username,
        token,
      });

      return {
        success: true,
        data: {
          token,
        },
      };
    },
  );

  fastify.post(
    "/api/v1/tunnels",
    {
      schema: {
        body: {
          type: "object",
          required: ["token"],
          properties: {
            token: { type: "string" },
          },
        },
      },
    },
    async (req, res) => {
      // @ts-expect-error
      const body: {
        token: string;
      } = req.body;

      const userData = instance.users.find(user => user.token == body.token);

      if (!userData)
        return res.status(403).send({
          error: "Invalid token",
        });

      // const host = req.hostname.substring(0, req.hostname.indexOf(":"));
      const unparsedPort = req.hostname.substring(
        req.hostname.indexOf(":") + 1,
      );

      // @ts-expect-error
      // parseInt(...) can take a number just fine, at least in Node.JS
      const port = parseInt(unparsedPort == "" ? proxiedPort : unparsedPort);

      // This protocol is so confusing. I'm sorry.
      res.send({
        success: true,
        data: instance.proxies.map(proxy => ({
          proxyUrlSettings: {
            host: "sameAs", // Makes pfC work (this is by design apparently)
            port,
            protocol: proxy.protocol.toUpperCase(),
          },

          dest: `${proxy.sourceIP}:${proxy.destPort}`,
          name: `${proxy.protocol.toUpperCase()} on ::${proxy.sourcePort} -> ::${proxy.destPort}`,

          passwords: [proxy.userConfig[userData.username]],

          running: true,
        })),
      });
    },
  );
}
