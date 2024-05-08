import type { Axios } from "axios";

import { SSHCommand } from "../libs/patchCommander.js";
import type { PrintLine } from "../commands.js";

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  apiKey: string,
) {
  if (argv.length == 1) return println("error: no arguments specified! run %s --help to see commands.\n", argv[0]);
  
  const program = new SSHCommand(println);
  program.description("Manages backends for NextNet");
  program.version("v0.1.0-preprod");

  const addBackend = new SSHCommand(println, "add");
  addBackend.description("Adds a backend");
  addBackend.argument("<name>", "Name of the backend");

  addBackend.argument(
    "<provider>",
    "Provider of the backend (ex. passyfire, ssh)",
  );

  addBackend.option(
    "-d, --description",
    "Description for the backend",
  );

  addBackend.option(
    "-f, --force-custom-parameters",
    "If turned on, this forces you to use custom parameters",
  );

  addBackend.option(
    "-c, --custom-parameters",
    "Custom parameters. Use this if the backend you're using isn't native to SSH yet, or if you manually turn on -f.",
  );

  // SSH provider
  addBackend.option(
    "-k, --ssh-key",
    "(SSH) SSH private key to use to authenticate with the server",
  );

  addBackend.option(
    "-u, --username",
    "(SSH, PassyFire) Username to authenticate with. With PassyFire, it's the username you create",
  );

  addBackend.option(
    "-h, --host",
    "(SSH, PassyFire) Host to connect to. With PassyFire, it's what you listen on",
  );

  // PassyFire provider
  addBackend.option(
    "-pe, --is-proxied",
    "(PassyFire) Specify if you're behind a proxy or not so we can get the right IP",
  );

  addBackend.option(
    "-pp, --proxied-port",
    "(PassyFire) If you're behind a proxy, and the port is different, specify the port to return",
  );

  addBackend.option("-g, --guest", "(PassyFire) Enable the guest user");

  addBackend.option(
    "-ua, --user-ask",
    "(PassyFire) Ask what users you want to create",
  );

  addBackend.option(
    "-p, --password",
    "(PassyFire) What password you want to use for the primary user",
  );

  const removeBackend = new SSHCommand(println, "rm");
  removeBackend.description("Removes a backend");
  removeBackend.argument("<id>", "Id of the backend");

  const lookupBackend = new SSHCommand(println, "find");
  lookupBackend.description("Looks up a backend based on your arguments");

  program.addCommand(addBackend);
  program.addCommand(removeBackend);
  program.addCommand(lookupBackend);

  program.parse(argv);
  await new Promise((resolve) => program.onExit(resolve));
}
