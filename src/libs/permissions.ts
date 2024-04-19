import type { PrismaClient } from "@prisma/client";

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

export async function hasPermission(permission: string, uid: number, prisma: PrismaClient): Promise<boolean> {
  const permissionNode = await prisma.permission.findFirst({
    where: {
      userID: uid,
      permission
    }
  });

  if (!permissionNode) return false;
  return permissionNode.has;
}