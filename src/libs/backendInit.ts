import type { PrismaClient } from "@prisma/client";

import type { BackendBaseClass } from "../backendimpl/base.js";
import { backendProviders } from "../backendimpl/index.js";

type Backend = {
  id: number;
  name: string; 
  description: string | null; 
  backend: string; 
  connectionDetails: string;
};

export async function backendInit(backend: Backend, backends: Record<number, BackendBaseClass>, prisma: PrismaClient): Promise<boolean> {
  const ourProvider = backendProviders[backend.backend];
  
  if (!ourProvider) {
    console.log(" - Error: Invalid backend recieved!");
    return false;
  }

  console.log(" - Initializing backend...");

  backends[backend.id] = new ourProvider(backend.connectionDetails);
  const ourBackend = backends[backend.id];

  if (!await ourBackend.start()) {
    console.log(" - Error initializing backend!");
    console.log("   - " + ourBackend.logs.join("\n   - "));

    return false;
  }

  console.log(" - Initializing clients...");

  const clients = await prisma.forwardRule.findMany({
    where: {
      destProviderID: backend.id,
      enabled: true
    }
  });

  for (const client of clients) {
    if (client.protocol != "tcp" && client.protocol != "udp") {
      console.error(` - Error: Client with ID of '${client.id}' has an invalid protocol! (must be either TCP or UDP)`);
      continue;
    }

    ourBackend.addConnection(client.sourceIP, client.sourcePort, client.destPort, client.protocol);
  }

  return true;
}