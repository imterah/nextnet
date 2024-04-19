/*
  Warnings:

  - You are about to drop the column `permissionID` on the `Permission` table. All the data in the column will be lost.
  - Added the required column `permission` to the `Permission` table without a default value. This is not possible if the table is not empty.

*/
-- RedefineTables
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_Permission" (
    "permission" TEXT NOT NULL,
    "has" BOOLEAN NOT NULL,
    "userID" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    CONSTRAINT "Permission_userID_fkey" FOREIGN KEY ("userID") REFERENCES "User" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);
INSERT INTO "new_Permission" ("has", "userID") SELECT "has", "userID" FROM "Permission";
DROP TABLE "Permission";
ALTER TABLE "new_Permission" RENAME TO "Permission";
PRAGMA foreign_key_check;
PRAGMA foreign_keys=ON;
