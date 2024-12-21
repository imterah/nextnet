-- CreateTable
CREATE TABLE "DesinationProvider" (
    "id" SERIAL NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "backend" TEXT NOT NULL,
    "connectionDetails" TEXT NOT NULL,

    CONSTRAINT "DesinationProvider_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ForwardRule" (
    "id" SERIAL NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "sourceIP" TEXT NOT NULL,
    "sourcePort" INTEGER NOT NULL,
    "destIP" TEXT NOT NULL,
    "destPort" INTEGER NOT NULL,
    "destProviderID" INTEGER NOT NULL,
    "enabled" BOOLEAN NOT NULL,

    CONSTRAINT "ForwardRule_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "Permission" (
    "id" SERIAL NOT NULL,
    "permission" TEXT NOT NULL,
    "has" BOOLEAN NOT NULL,
    "userID" INTEGER NOT NULL,

    CONSTRAINT "Permission_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "User" (
    "id" SERIAL NOT NULL,
    "email" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "rootToken" TEXT,
    "isRootServiceAccount" BOOLEAN,

    CONSTRAINT "User_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "User_email_key" ON "User"("email");

-- AddForeignKey
ALTER TABLE "Permission" ADD CONSTRAINT "Permission_userID_fkey" FOREIGN KEY ("userID") REFERENCES "User"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
