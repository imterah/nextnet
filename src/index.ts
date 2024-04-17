import process from "node:process";

import { PrismaClient } from '@prisma/client';
import Fastify from "fastify";

const prisma = new PrismaClient();

const fastify = Fastify({
  logger: true
});

fastify.get('/', async function handler (request, reply) {
  return { hello: 'world' };
})

// Run the server!
try {
  await fastify.listen({ port: 3000 });
} catch (err) {
  fastify.log.error(err);
  process.exit(1);
}