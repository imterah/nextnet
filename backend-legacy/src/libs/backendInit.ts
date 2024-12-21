import { format } from "node:util";

import type { PrismaClient } from "@prisma/client";

import { backendProviders } from "../backendimpl/index.js";
import { BackendBaseClass } from "../backendimpl/base.js";

type Backend = {
  id: number;
  name: string;
  description: string | null;
  backend: string;
  connectionDetails: string;
};

export async function backendInit(
  backend: Backend,
  backends: Record<number, BackendBaseClass>,
  prisma: PrismaClient,
  logger?: (arg: string) => void,
  errorOut?: (arg: string) => void,
): Promise<boolean> {
  const log = (...args: string[]) =>
    logger ? logger(format(...args)) : console.log(...args);

  const error = (...args: string[]) =>
    errorOut ? errorOut(format(...args)) : log(...args);

  const ourProvider = backendProviders[backend.backend];

  if (!ourProvider) {
    error(" - Error: Invalid backend recieved!");

    // Prevent crashes when we don't recieve a backend
    backends[backend.id] = new BackendBaseClass("");

    backends[backend.id].logs.push("** Failed To Create Backend **");

    backends[backend.id].logs.push(
      "Reason: Invalid backend recieved (couldn't find the backend to use!)",
    );

    return false;
  }

  log(" - Initializing backend...");

  backends[backend.id] = new ourProvider(backend.connectionDetails);
  const ourBackend = backends[backend.id];

  if (!(await ourBackend.start())) {
    error(" - Error initializing backend!");
    error("   - " + ourBackend.logs.join("\n   - "));

    return false;
  }

  log(" - Initializing clients...");

  const clients = await prisma.forwardRule.findMany({
    where: {
      destProviderID: backend.id,
      enabled: true,
    },
  });

  for (const client of clients) {
    if (client.protocol != "tcp" && client.protocol != "udp") {
      error(
        ` - Error: Client with ID of '${client.id}' has an invalid protocol! (must be either TCP or UDP)`,
      );
      continue;
    }

    ourBackend.addConnection(
      client.sourceIP,
      client.sourcePort,
      client.destPort,
      client.protocol,
    );
  }

  return true;
}
