import { createReadStream, createWriteStream } from "node:fs";
import { Readable, pipeline } from "node:stream";
import { createGzip } from "node:zlib";
import process from "node:process";

import { PrismaClient } from "@prisma/client";

const gzip = createGzip();

if (process.argv.length <= 2) {
  console.error(
    "Missing arguments! Usage: node ./out/tools/exportDBContents.js exportPath.json.gz",
  );
  process.exit(1);
}

console.log("Initializing Database...");
const prisma = new PrismaClient();
console.log("Initialized Database.");

console.log("Getting all destinationProviders...");
const destinationProviders = await prisma.desinationProvider.findMany();

console.log("Getting all forwardRules...");
const forwardRules = await prisma.forwardRule.findMany();

console.log("Getting all permissions...");
const allPermissions = await prisma.permission.findMany();

console.log("Getting all users...");
const users = await prisma.user.findMany();

const masterList = JSON.stringify({
  destinationProviders,
  forwardRules,
  allPermissions,
});

const source = new Readable();
source.push(masterList);
source.push(null);

const destination = createWriteStream(process.argv[2]);

pipeline(source, gzip, destination, err => {
  if (err) {
    console.error("Failed to compress JSON data:", err);
  } else {
    console.log("Sucesfully saved DB contents.");
  }
});
