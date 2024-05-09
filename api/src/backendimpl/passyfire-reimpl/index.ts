import fastifyWebsocket from "@fastify/websocket";

import type { FastifyInstance } from "fastify";
import Fastify from "fastify";

import type {
  ForwardRule,
  ConnectedClient,
  ParameterReturnedValue,
  BackendBaseClass,
} from "../base.js";

import { generateRandomData } from "../../libs/generateRandom.js";
import { requestHandler } from "./socket.js";
import { route } from "./routes.js";

type BackendProviderUser = {
  username: string;
  password: string;
};

export type ForwardRuleExt = ForwardRule & {
  protocol: "tcp" | "udp";
  userConfig: Record<string, string>;
};

export type ConnectedClientExt = ConnectedClient & {
  connectionDetails: ForwardRuleExt;
  username: string;
};

// Fight me (for better naming)
type BackendParsedProviderString = {
  ip: string;
  port: number;
  publicPort?: number;
  isProxied?: boolean;

  users: BackendProviderUser[];
};

type LoggedInUser = {
  username: string;
  token: string;
};

function parseBackendProviderString(data: string): BackendParsedProviderString {
  try {
    JSON.parse(data);
  } catch (e) {
    throw new Error("Payload body is not JSON");
  }

  const jsonData = JSON.parse(data);

  if (typeof jsonData.ip != "string")
    throw new Error("IP field is not a string");
  
  if (typeof jsonData.port != "number") throw new Error("Port is not a number");

  if (
    typeof jsonData.publicPort != "undefined" &&
    typeof jsonData.publicPort != "number"
  )
    throw new Error("(optional field) Proxied port is not a number");
  
  if (
    typeof jsonData.isProxied != "undefined" &&
    typeof jsonData.isProxied != "boolean"
  )
    throw new Error("(optional field) 'Is proxied' is not a boolean");

  if (!Array.isArray(jsonData.users)) throw new Error("Users is not an array");

  for (const userIndex in jsonData.users) {
    const user = jsonData.users[userIndex];

    if (typeof user.username != "string")
      throw new Error("Username is not a string, in users array");
    if (typeof user.password != "string")
      throw new Error("Password is not a string, in users array");
  }

  return {
    ip: jsonData.ip,
    port: jsonData.port,

    publicPort: jsonData.publicPort,
    isProxied: jsonData.isProxied,

    users: jsonData.users,
  };
}

export class PassyFireBackendProvider implements BackendBaseClass {
  state: "stopped" | "stopping" | "started" | "starting";

  clients: ConnectedClientExt[];
  proxies: ForwardRuleExt[];
  users: LoggedInUser[];
  logs: string[];

  options: BackendParsedProviderString;
  fastify: FastifyInstance;

  constructor(parameters: string) {
    this.logs = [];
    this.clients = [];
    this.proxies = [];

    this.state = "stopped";
    this.options = parseBackendProviderString(parameters);

    this.users = [];
  }

  async start(): Promise<boolean> {
    this.state = "starting";

    this.fastify = Fastify({
      logger: true,
      trustProxy: this.options.isProxied,
    });

    await this.fastify.register(fastifyWebsocket);
    route(this);

    this.fastify.get("/", { websocket: true }, (ws, req) =>
      requestHandler(this, ws, req),
    );

    await this.fastify.listen({
      port: this.options.port,
      host: this.options.ip,
    });

    this.state = "started";

    return true;
  }

  async stop(): Promise<boolean> {
    await this.fastify.close();

    this.users.splice(0, this.users.length);
    this.proxies.splice(0, this.proxies.length);
    this.clients.splice(0, this.clients.length);

    return true;
  }

  addConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const proxy: ForwardRuleExt = {
      sourceIP,
      sourcePort,
      destPort,
      protocol,

      userConfig: {},
    };

    for (const user of this.options.users) {
      proxy.userConfig[user.username] = generateRandomData();
    }

    this.proxies.push(proxy);
  }

  removeConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const connectionCheck = PassyFireBackendProvider.checkParametersConnection(
      sourceIP,
      sourcePort,
      destPort,
      protocol,
    );
    if (!connectionCheck.success) throw new Error(connectionCheck.message);

    const foundProxyEntry = this.proxies.find(
      i =>
        i.sourceIP == sourceIP &&
        i.sourcePort == sourcePort &&
        i.destPort == destPort,
    );
    if (!foundProxyEntry) return;

    this.proxies.splice(this.proxies.indexOf(foundProxyEntry), 1);
    return;
  }

  getAllConnections(): ConnectedClient[] {
    if (this.clients == null) return [];
    return this.clients;
  }

  static checkParametersConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): ParameterReturnedValue {
    return {
      success: true,
    };
  }

  static checkParametersBackendInstance(data: string): ParameterReturnedValue {
    try {
      parseBackendProviderString(data);
      // @ts-ignore
    } catch (e: Error) {
      return {
        success: false,
        message: e.toString(),
      };
    }

    return {
      success: true,
    };
  }
}
