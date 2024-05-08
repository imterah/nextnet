import type { Axios } from "axios";

import { SSHCommand } from "../libs/patchCommander.js";
import type { PrintLine } from "../commands.js";

// https://stackoverflow.com/questions/37938504/what-is-the-best-way-to-find-all-items-are-deleted-inserted-from-original-arra
function difference(a: any[], b: any[]) {
	return a.filter(x => b.indexOf(x) < 0);
};

type InboundConnectionSuccess = {
  success: true,
  data: {
    ip: string,
    port: number,

    connectionDetails: {
      sourceIP: string,
      sourcePort: number,
      destPort: number,
      enabled: boolean
    }
  }[]
};

type LookupCommandSuccess = {
  success: true,
  data: {
    id: number,
    name: string,
    description: string,
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    providerID: number,
    autoStart: boolean
  }[]
};

export async function run(
  argv: string[],
  println: PrintLine,
  axios: Axios,
  token: string,
) {
  if (argv.length == 1) return println("error: no arguments specified! run %s --help to see commands.\n", argv[0]);

  const program = new SSHCommand(println);
  program.description("Manages connections for NextNet");
  program.version("v1.0.0-testing");

  const addCommand = new SSHCommand(println, "add");
  addCommand.description("Creates a new connection");

  addCommand.argument(
    "<backend_id>",
    "The backend ID to use. Can be fetched by the command 'backend search'",
  );

  addCommand.argument("<name>", "The name for the tunnel");
  addCommand.argument("<protocol>", "The protocol to use. Either TCP or UDP");

  addCommand.argument(
    "<source>",
    "Source IP and port combo (ex. '192.168.0.63:25565'",
  );

  addCommand.argument("<dest_port>", "Destination port to use");
  addCommand.option("-d, --description", "Description for the tunnel");

  addCommand.action(async(providerIDStr: string, name: string, protocolRaw: string, source: string, destPortRaw: string, options: {
    description?: string
  }) => {
    const providerID = parseInt(providerIDStr);

    if (Number.isNaN(providerID)) {
      println("ID (%s) is not a number\n", providerIDStr);
      return;
    };

    const protocol = protocolRaw.toLowerCase().trim();

    if (protocol != "tcp" && protocol != "udp") {
      return println("Protocol is not a valid option (not tcp or udp)\n");
    };

    const sourceSplit: string[] = source.split(":");

    if (sourceSplit.length != 2) {
      return println("Source could not be splitted down (are you missing the ':' in the source to specify port?)\n");
    }

    const sourceIP: string = sourceSplit[0];
    const sourcePort: number = parseInt(sourceSplit[1]);

    if (Number.isNaN(sourcePort)) {
      return println("Port splitted is not a number\n");
    }

    const destinationPort: number = parseInt(destPortRaw);

    if (Number.isNaN(destinationPort)) {
      return println("Destination port could not be parsed into a number\n");
    }

    const response = await axios.post("/api/v1/forward/create", {
      token,

      name,
      description: options.description,

      protocol,

      sourceIP,
      sourcePort,

      destinationPort,

      providerID
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error requesting connections!\n");
      }

      return;
    }

    println("Successfully created connection.\n");
  });

  const lookupCommand = new SSHCommand(println, "find");

  lookupCommand.description(
    "Looks up all connections based on the arguments you specify",
  );

  lookupCommand.option(
    "-b, --backend-id <id>",
    "The backend ID to use. Can be fetched by 'back find'",
  );

  lookupCommand.option("-n, --name <name>", "The name for the tunnel");

  lookupCommand.option(
    "-p, --protocol <protocol>",
    "The protocol to use. Either TCP or UDP",
  );

  lookupCommand.option(
    "-s <source>, --source",
    "Source IP and port combo (ex. '192.168.0.63:25565'",
  );

  lookupCommand.option("-d, --dest-port <port>", "Destination port to use");

  lookupCommand.option(
    "-o, --description <description>",
    "Description for the tunnel",
  );

  lookupCommand.action(async(options: {
    backendId?: string,
    destPort?: string,
    name?: string,
    protocol?: string,
    source?: string,
    description?: string
  }) => {
    let numberBackendID: number | undefined;

    let sourceIP:   string | undefined;
    let sourcePort: number | undefined;

    let destPort:   number | undefined;

    if (options.backendId) {
      numberBackendID = parseInt(options.backendId);

      if (Number.isNaN(numberBackendID)) {
        println("ID (%s) is not a number\n", options.backendId);
        return;
      }
    }

    if (options.source) {
      const sourceSplit: string[] = options.source.split(":");

      if (sourceSplit.length != 2) {
        return println("Source could not be splitted down (are you missing the ':' in the source to specify port?)\n");
      }
  
      sourceIP = sourceSplit[0];
      sourcePort = parseInt(sourceSplit[1]);
  
      if (Number.isNaN(sourcePort)) {
        return println("Port splitted is not a number\n");
      }
    }

    if (options.destPort) {
      destPort = parseInt(options.destPort);

      if (Number.isNaN(destPort)) {
        println("ID (%s) is not a number\n", options.destPort);
        return;
      }
    }

    const response = await axios.post("/api/v1/forward/lookup", {
      token,

      name: options.name,
      description: options.description,

      protocol: options.protocol,

      sourceIP,
      sourcePort,

      destinationPort: destPort
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error requesting connections!\n");
      }

      return;
    }

    const { data }: LookupCommandSuccess = response.data;

    for (const connection of data) {
      println("ID: %s%s:\n", connection.id, (connection.autoStart ? " (automatically starts)" : ""));
      println(" - Backend ID: %s\n", connection.providerID);
      println(" - Name: %s\n", connection.name);
      if (connection.description) println(" - Description: %s\n", connection.description);
      println(" - Source: %s:%s\n", connection.sourceIP, connection.sourcePort);
      println(" - Destination port: %s\n", connection.destPort);

      println("\n");
    }

    println("%s connections found.\n", data.length);
  });

  const startTunnel = new SSHCommand(println, "start");
  startTunnel.description("Starts a tunnel");
  startTunnel.argument("<id>", "Tunnel ID to start");

  startTunnel.action(async(idStr: string) => {
    const id = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number\n", idStr);
      return;
    };

    const response = await axios.post("/api/v1/forward/start", {
      token,
      id
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error starting the connection!\n");
      }

      return;
    }

    println("Successfully started tunnel.\n");
    return;
  });

  const stopTunnel = new SSHCommand(println, "stop");
  stopTunnel.description("Stops a tunnel");
  stopTunnel.argument("<id>", "Tunnel ID to stop");

  stopTunnel.action(async(idStr: string) => {
    const id = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number\n", idStr);
      return;
    };

    const response = await axios.post("/api/v1/forward/stop", {
      token,
      id
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error stopping a connection!\n");
      }

      return;
    }

    println("Successfully stopped tunnel.\n");
  });

  const getInbound = new SSHCommand(println, "get-inbound");
  getInbound.description("Shows all current connections");
  getInbound.argument("<id>", "Tunnel ID to view inbound connections of");
  getInbound.option("-t, --tail", "Live-view of connection list");
  getInbound.option("-s, --tail-pull-rate <ms>", "Controls the speed to pull at (in ms)");

  getInbound.action(async(idStr: string, options: {
    tail?: boolean,
    tailPullRate?: string
  }): Promise<void> => {  
    const pullRate: number = options.tailPullRate ? parseInt(options.tailPullRate) : 2000;
    const id = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number\n", idStr);
      return;
    }

    if (Number.isNaN(pullRate)) {
      println("Pull rate is not a number\n");
      return;
    }

    if (options.tail) {
      let previousEntries: string[] = [];

      while (true) {
        const response = await axios.post("/api/v1/forward/connections", {
          token,
          id
        });

        if (response.status != 200) {
          if (process.env.NODE_ENV != "production") console.log(response);
  
          if (response.data.error) {
            println(`Error: ${response.data.error}\n`);
          } else {
            println("Error requesting connections!\n");
          }
  
          return;
        }
  
        const { data }: InboundConnectionSuccess = response.data;
        const simplifiedArray: string[] = data.map((i) => `${i.ip}:${i.port}`);

        const insertedItems: string[] = difference(simplifiedArray, previousEntries);
        const removedItems: string[] = difference(previousEntries, simplifiedArray);

        insertedItems.forEach((i) => println("CONNECTED:    %s\n", i));
        removedItems.forEach((i) => println("DISCONNECTED: %s\n", i));

        previousEntries = simplifiedArray;

        await new Promise((i) => setTimeout(i, pullRate));
      }
    } else {
      const response = await axios.post("/api/v1/forward/connections", {
        token,
        id
      });

      if (response.status != 200) {
        if (process.env.NODE_ENV != "production") console.log(response);

        if (response.data.error) {
          println(`Error: ${response.data.error}\n`);
        } else {
          println("Error requesting connections!\n");
        }

        return;
      }

      const { data }: InboundConnectionSuccess = response.data;

      if (data.length == 0) {
        println("There are currently no connected clients.\n");
        return;
      }

      println("Connected clients (for source: %s:%s):\n", data[0].connectionDetails.sourceIP, data[0].connectionDetails.sourcePort);

      for (const entry of data) {
        println(" - %s:%s\n", entry.ip, entry.port);
      }
    }
  });

  const removeTunnel = new SSHCommand(println, "rm");
  removeTunnel.description("Removes a tunnel");
  removeTunnel.argument("<id>", "Tunnel ID to remove");

  removeTunnel.action(async(idStr: string) => {
    const id = parseInt(idStr);

    if (Number.isNaN(id)) {
      println("ID (%s) is not a number\n", idStr);
      return;
    };

    const response = await axios.post("/api/v1/forward/remove", {
      token,
      id
    });

    if (response.status != 200) {
      if (process.env.NODE_ENV != "production") console.log(response);

      if (response.data.error) {
        println(`Error: ${response.data.error}\n`);
      } else {
        println("Error deleting connection!\n");
      }

      return;
    };

    println("Successfully deleted connection.\n");
  });

  program.addCommand(addCommand);
  program.addCommand(lookupCommand);
  program.addCommand(startTunnel);
  program.addCommand(stopTunnel);
  program.addCommand(getInbound);
  program.addCommand(removeTunnel);

  program.parse(argv);
  await new Promise((resolve) => program.onExit(resolve));
}
