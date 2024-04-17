-- CreateTable
CREATE TABLE "DesinationProvider" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "backend" TEXT NOT NULL,
    "connectionDetails" TEXT NOT NULL
);

-- CreateTable
CREATE TABLE "ForwardRule" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "sourceIP" TEXT NOT NULL,
    "sourcePort" INTEGER NOT NULL,
    "destIP" TEXT NOT NULL,
    "destPort" INTEGER NOT NULL,
    "destProviderID" INTEGER NOT NULL,
    "enabled" BOOLEAN NOT NULL
);

-- CreateTable
CREATE TABLE "Permission" (
    "permissionID" TEXT NOT NULL,
    "has" BOOLEAN NOT NULL,
    "userID" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    CONSTRAINT "Permission_userID_fkey" FOREIGN KEY ("userID") REFERENCES "User" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "User" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "email" TEXT NOT NULL,
    "name" TEXT,
    "rootToken" TEXT
);

-- CreateIndex
CREATE UNIQUE INDEX "Permission_permissionID_key" ON "Permission"("permissionID");

-- CreateIndex
CREATE UNIQUE INDEX "User_email_key" ON "User"("email");
