import type { Axios } from "axios";

import { SSHCommand } from "../libs/patchCommander.js";
import type { PrintLine } from "../commands.js";

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
) {
  const program = new SSHCommand(println);
  program.description("Manages users for NextNet");
  program.version("v0.1.0-preprod");

  const addCommand = new SSHCommand(println, "add");
  addCommand.description("Create a new user");
  addCommand.argument("<username>", "Username of new user");
  addCommand.argument("<email>", "Email of new user");
  addCommand.argument("[name]", "Name of new user (defaults to username)");

  addCommand.option("-p, --password", "Password of User");
  addCommand.option(
    "-a, --ask-password, --ask-pass, --askpass",
    "Asks for a password. Hides output",
  );

  const removeCommand = new SSHCommand(println, "rm");
  removeCommand.description("Remove a user");
  removeCommand.argument("<uid>", "ID of user to remove");

  const lookupCommand = new SSHCommand(println, "find");
  lookupCommand.description("Find a user");
  lookupCommand.option("-i, --id <id>", "UID of User");
  lookupCommand.option("-n, --name <name>", "Name of User");
  lookupCommand.option("-u, --username <username>", "Username of User");
  lookupCommand.option("-e, --email <email>", "Email of User");
  lookupCommand.option("-s, --service", "The User is a service account");

  program.addCommand(addCommand);
  program.addCommand(removeCommand);
  program.addCommand(lookupCommand);

  program.parse(argv);
}
