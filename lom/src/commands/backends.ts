import type { Axios } from "axios";

import { SSHCommand } from "../libs/patchCommander.js";
import type { PrintLine, KeyboardRead } from "../commands.js";

type BackendLookupSuccess = {
  success: boolean;
  data: {
    id: number;

    name: string;
    description: string;
    backend: string;
    connectionDetails?: string;
    logs: string[];
  }[];
};

const addRequiredOptions = {
  ssh: ["sshKey", "username", "host"],
  sshpy: ["sshKey", "username", "host"],

  passyfire: ["host"],
};

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  token: string,
  readKeyboard: KeyboardRead,
) {
  const program = new SSHCommand(println);
  program.description("Manages backends for NextNet");
  program.version("v1.0.0");

  const addBackend = new SSHCommand(println, "add");

  addBackend.description("Adds a backend");
  addBackend.argument("<name>", "Name of the backend");

  addBackend.argument(
    "<provider>",
    "Provider of the backend (ex. passyfire, ssh)",
  );

  addBackend.option(
    "-d, --description <description>",
    "Description for the backend",
  );

  addBackend.option(
    "-f, --force-custom-parameters",
    "If turned on, this forces you to use custom parameters",
  );

  addBackend.option(
    "-c, --custom-parameters <parameters>",
    "Custom parameters. Use this if the backend you're using isn't native to SSH yet, or if you manually turn on -f.",
  );

  // SSH provider
  addBackend.option(
    "-k, --ssh-key <private-key>",
    "(SSH, SSHpy) SSH private key to use to authenticate with the server",
  );

  addBackend.option(
    "-u, --username <user>",
    "(SSH, SSHpy, PassyFire) Username to authenticate with. With PassyFire, it's the username you create",
  );

  addBackend.option(
    "-h, --host <host>",
    "(SSH, SSHpy, PassyFire) Host to connect to. With PassyFire, it's what you listen on",
  );

  // PassyFire provider
  addBackend.option(
    "-pe, --is-proxied",
    "(PassyFire) Specify if you're behind a proxy or not so we can get the right IP",
  );

  addBackend.option(
    "-pp, --proxied-port <port>",
    "(PassyFire) If you're behind a proxy, and the port is different, specify the port to return",
  );

  addBackend.option("-g, --guest", "(PassyFire) Enable the guest user");

  addBackend.option(
    "-ua, --user-ask",
    "(PassyFire) Ask what users you want to create",
  );

  addBackend.option(
    "-p, --password <password>",
    "(PassyFire) What password you want to use for the primary user",
  );

  addBackend.action(
    async (
      name: string,
      provider: string,
      options: {
        description?: string;
        forceCustomParameters?: boolean;
        customParameters?: string;

        // SSH (mostly)
        sshKey?: string;
        username?: string;
        host?: string;

        // PassyFire (mostly)
        isProxied?: boolean;
        proxiedPort?: string;
        guest?: boolean;
        userAsk?: boolean;
        password?: string;
      },
    ) => {
      // @ts-expect-error: Yes it can index for what we need it to do.
      const isUnsupportedPlatform: boolean = !addRequiredOptions[provider];

      if (isUnsupportedPlatform) {
        println(
          "WARNING: Platform is not natively supported by the LOM yet!\n",
        );
      }

      let connectionDetails: string = "";

      if (options.forceCustomParameters || isUnsupportedPlatform) {
        if (typeof options.customParameters != "string") {
          return println(
            "ERROR: You are missing the custom parameters option!\n",
          );
        }

        connectionDetails = options.customParameters;
      } else if (provider == "ssh") {
        for (const argument of addRequiredOptions["ssh"]) {
          // @ts-expect-error: No.
          const hasArgument = options[argument];

          if (!hasArgument) {
            return println("ERROR: Missing argument '%s'\n", argument);
          }
        }

        const unstringifiedArguments: {
          ip?: string;
          port?: number;
          username?: string;
          privateKey?: string;
        } = {};

        if (options.host) {
          const sourceSplit: string[] = options.host.split(":");

          const sourceIP: string = sourceSplit[0];
          const sourcePort: number =
            sourceSplit.length >= 2 ? parseInt(sourceSplit[1]) : 22;

          unstringifiedArguments.ip = sourceIP;
          unstringifiedArguments.port = sourcePort;
        }

        unstringifiedArguments.username = options.username;
        unstringifiedArguments.privateKey = options.sshKey?.replaceAll(
          "\\n",
          "\n",
        );

        connectionDetails = JSON.stringify(unstringifiedArguments);
      } else if (provider == "sshpy") {
        // TODO add full functionality
        for (const argument of addRequiredOptions["ssh"]) {
          // @ts-expect-error: No.
          const hasArgument = options[argument];

          if (!hasArgument) {
            return println("ERROR: Missing argument '%s'\n", argument);
          }
        }

        const unstringifiedArguments: {
          ip?: string;
          port?: number;
          username?: string;
          privateKey?: string;
        } = {};

        if (options.host) {
          const sourceSplit: string[] = options.host.split(":");

          const sourceIP: string = sourceSplit[0];
          const sourcePort: number =
            sourceSplit.length >= 2 ? parseInt(sourceSplit[1]) : 22;

          unstringifiedArguments.ip = sourceIP;
          unstringifiedArguments.port = sourcePort;
        }

        unstringifiedArguments.username = options.username;

        unstringifiedArguments.privateKey = options.sshKey?.replaceAll(
          "\\n",
          "\n",
        );

        console.log(unstringifiedArguments.privateKey?.indexOf("\n"));

        connectionDetails = JSON.stringify(unstringifiedArguments);
      } else if (provider == "passyfire") {
        for (const argument of addRequiredOptions["passyfire"]) {
          // @ts-expect-error: No.
          const hasArgument = options[argument];

          if (!hasArgument) {
            return println("ERROR: Missing argument '%s'\n", argument);
          }
        }

        const unstringifiedArguments: {
          ip?: string;
          port?: number;
          publicPort?: number;
          isProxied?: boolean;
          users: {
            username: string;
            password: string;
          }[];
        } = {
          users: [],
        };

        if (options.guest) {
          unstringifiedArguments.users.push({
            username: "guest",
            password: "guest",
          });
        }

        if (options.username) {
          if (!options.password) {
            return println("Password must not be left blank\n");
          }

          unstringifiedArguments.users.push({
            username: options.username,
            password: options.password,
          });
        }

        if (options.userAsk) {
          let shouldContinueAsking: boolean = true;

          while (shouldContinueAsking) {
            println("Creating a user.\nUsername: ");
            const username = await readKeyboard();

            let passwordConfirmOne = "a";
            let passwordConfirmTwo = "b";

            println("\n");

            while (passwordConfirmOne != passwordConfirmTwo) {
              println("Password: ");
              passwordConfirmOne = await readKeyboard(true);

              println("\nConfirm password: ");
              passwordConfirmTwo = await readKeyboard(true);

              println("\n");

              if (passwordConfirmOne != passwordConfirmTwo) {
                println("Passwords do not match! Try again.\n\n");
              }
            }

            unstringifiedArguments.users.push({
              username,
              password: passwordConfirmOne,
            });

            println("\nShould we continue creating users? (y/n) ");
            shouldContinueAsking = (await readKeyboard())
              .toLowerCase()
              .trim()
              .startsWith("y");

            println("\n\n");
          }
        }

        if (unstringifiedArguments.users.length == 0) {
          return println(
            "No users will be created with your current arguments! You must have users set up.\n",
          );
        }

        unstringifiedArguments.isProxied = Boolean(options.isProxied);

        if (options.proxiedPort) {
          unstringifiedArguments.publicPort = parseInt(
            options.proxiedPort ?? "",
          );

          if (Number.isNaN(unstringifiedArguments.publicPort)) {
            println("UID (%s) is not a number.\n", options.proxiedPort);
            return;
          }
        }

        if (options.host) {
          const sourceSplit: string[] = options.host.split(":");

          if (sourceSplit.length != 2) {
            return println(
              "Source could not be splitted down (are you missing the ':' in the source to specify port?)\n",
            );
          }

          const sourceIP: string = sourceSplit[0];
          const sourcePort: number = parseInt(sourceSplit[1]);

          if (Number.isNaN(sourcePort)) {
            println("UID (%s) is not a number.\n", sourcePort);
            return;
          }

          unstringifiedArguments.ip = sourceIP;
          unstringifiedArguments.port = sourcePort;
        }

        connectionDetails = JSON.stringify(unstringifiedArguments);
      }

      const response = await axios.post("/api/v1/backends/create", {
        token,

        name,
        description: options.description,
        backend: provider,

        connectionDetails,
      });

      if (response.status != 200) {
        if (process.env.NODE_ENV != "production") console.log(response);

        if (response.data.error) {
          println(`Error: ${response.data.error}\n`);
        } else {
          println("Error creating a backend!\n");
        }

        return;
      }

      println("Successfully created the backend.\n");
    },
  );

  const removeBackend = new SSHCommand(println, "rm");
  removeBackend.description("Removes a backend");
  removeBackend.argument("<id>", "ID of the backend");

  removeBackend.action(async (idStr: string) => {
    const id: number = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number.\n", idStr);
      return;
    }

    const response = await axios.post("/api/v1/backends/remove", {
      token,
      id,
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error deleting backend!\n");
      }

      return;
    }

    println("Backend has been successfully deleted.\n");
  });

  const lookupBackend = new SSHCommand(println, "find");
  lookupBackend.description("Looks up a backend based on your arguments");

  lookupBackend.option("-n, --name <name>", "Name of the backend");

  lookupBackend.option(
    "-p, --provider <provider>",
    "Provider of the backend (ex. passyfire, ssh)",
  );

  lookupBackend.option(
    "-d, --description <description>",
    "Description for the backend",
  );

  lookupBackend.option(
    "-e, --parse-connection-details",
    "If specified, we automatically parse the connection details to make them human readable, if standard JSON.",
  );

  lookupBackend.action(
    async (options: {
      name?: string;
      provider?: string;
      description?: string;
      parseConnectionDetails?: boolean;
    }) => {
      const response = await axios.post("/api/v1/backends/lookup", {
        token,

        name: options.name,
        description: options.description,

        backend: options.provider,
      });

      if (response.status != 200) {
        if (process.env.NODE_ENV != "production") console.log(response);

        if (response.data.error) {
          println(`Error: ${response.data.error}\n`);
        } else {
          println("Error looking up backends!\n");
        }

        return;
      }

      const { data }: BackendLookupSuccess = response.data;

      for (const backend of data) {
        println("ID: %s:\n", backend.id);
        println(" - Name: %s\n", backend.name);
        println(" - Description: %s\n", backend.description);
        println(" - Using Backend: %s\n", backend.backend);

        if (backend.connectionDetails) {
          if (options.parseConnectionDetails) {
            // We don't know what we're recieving. We just try to parse it (hence the any type)
            // {} is more accurate but TS yells at us if we do that :(

            // eslint-disable-next-line
            let parsedJSONData: any | undefined;

            try {
              parsedJSONData = JSON.parse(backend.connectionDetails);
            } catch (e) {
              println(" - Connection Details: %s\n", backend.connectionDetails);
              continue;
            }

            if (!parsedJSONData) {
              // Not really an assertion but I don't care right now
              println(
                "Assertion failed: parsedJSONData should not be undefined\n",
              );
              continue;
            }

            println(" - Connection details:\n");

            for (const key of Object.keys(parsedJSONData)) {
              let value: string | number = parsedJSONData[key];

              if (typeof value == "string") {
                value = value.replaceAll("\n", "\n" + " ".repeat(16));
              }

              if (typeof value == "object") {
                // TODO: implement?
                value = JSON.stringify(value);
              }

              println("  - %s: %s\n", key, value);
            }
          } else {
            println(" - Connection Details: %s\n", backend.connectionDetails);
          }
        }

        println("\n");
      }

      println("%s backends found.\n", data.length);
    },
  );

  const logsCommand = new SSHCommand(println, "logs");
  logsCommand.description("View logs for a backend");
  logsCommand.argument("<id>", "ID of the backend");

  logsCommand.action(async (idStr: string) => {
    const id: number = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number.\n", idStr);
      return;
    }

    const response = await axios.post("/api/v1/backends/lookup", {
      token,
      id,
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error getting logs!\n");
      }

      return;
    }

    const { data }: BackendLookupSuccess = response.data;
    const ourBackend = data.find(i => i.id == id);

    if (!ourBackend) return println("Could not find the backend!\n");
    ourBackend.logs.forEach(log => println("%s\n", log));
  });

  program.addCommand(addBackend);
  program.addCommand(removeBackend);
  program.addCommand(lookupBackend);
  program.addCommand(logsCommand);

  program.parse(argv);

  // It would make sense to check this, then parse argv, however this causes issues with
  // the application name not displaying correctly.

  if (argv.length == 1) {
    println("No arguments specified!\n\n");
    program.help();
    return;
  }

  await new Promise(resolve => program.onExit(resolve));
}
