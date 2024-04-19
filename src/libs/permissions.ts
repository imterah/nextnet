import type { PrismaClient } from "@prisma/client";
import type { SessionToken } from "./types.js";

export const permissionListDisabled: Record<string, boolean> = {
  "routes.add":         false,
  "routes.remove":      false,
  "routes.start":       false,
  "routes.stop":        false,
  "routes.edit":        false,
  "routes.visible":     false,

  "backends.add":       false,  
  "backends.remove":    false,
  "backends.start":     false,
  "backends.stop":      false,
  "backends.edit":      false,
  "backends.visible":   false,
  "backends.secretVis": false,

  "permissions.see":    false, 

  "users.add":          false,
  "users.remove":       false
};

// FIXME: This solution fucking sucks.
export let permissionListEnabled: Record<string, boolean> = JSON.parse(JSON.stringify(permissionListDisabled));

for (const index of Object.keys(permissionListEnabled)) {
  permissionListEnabled[index] = true;
}

export async function hasPermission(permissionList: string[], uid: number, prisma: PrismaClient): Promise<boolean> {
  for (const permission of permissionList) {
    const permissionNode = await prisma.permission.findFirst({
      where: {
        userID: uid,
        permission
      }
    });

    if (!permissionNode || !permissionNode.has) return false;
  }

  return true;
}

export async function hasPermissionByToken(permissionList: string[], token: string, tokens: Record<number, SessionToken[]>, prisma: PrismaClient): Promise<boolean> {
  let userID = -1;

  // Look up in our currently authenticated users
  for (const otherTokenKey of Object.keys(tokens)) {
    const otherTokenList = tokens[parseInt(otherTokenKey)];

    for (const otherTokenIndex in otherTokenList) {
      const otherToken = otherTokenList[otherTokenIndex];

      if (otherToken.token == token) {
        if (otherToken.expiresAt < otherToken.createdAt + (otherToken.createdAt - Date.now())) {
          otherTokenList.splice(parseInt(otherTokenIndex), 1);
          continue;
        } else {
          userID = parseInt(otherTokenKey);
        }
      }
    }
  }

  // Fine, we'll look up for global tokens...
  // FIXME: Could this be more efficient? IDs are sequential in SQL I think
  if (userID == -1) {
    const allUsers = await prisma.user.findMany();
  
    for (const user of allUsers) {
      if (user.rootToken == token) userID = user.id;
    };
  }

  // If we are STILL -1, we give up.
  if (userID == -1) return false;

  // Now we can test permissions!
  return await hasPermission(permissionList, userID, prisma);
}