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
  program.description("Manages connections for NextNet");
  program.version("v0.1.0-preprod");

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

  addCommand.option("--description, -d", "Description for the tunnel");

  const lookupCommand = new SSHCommand(println, "find");

  lookupCommand.description(
    "Looks up all connections based on the arguments you specify",
  );

  lookupCommand.option(
    "--backend_id, -b <id>",
    "The backend ID to use. Can be fetched by 'back find'",
  );

  lookupCommand.option("--name, -n <name>", "The name for the tunnel");

  lookupCommand.option(
    "--protocol, -p <protocol>",
    "The protocol to use. Either TCP or UDP",
  );

  lookupCommand.option(
    "--source, -s <source>",
    "Source IP and port combo (ex. '192.168.0.63:25565'",
  );

  lookupCommand.option("--dest_port, -d <port>", "Destination port to use");

  lookupCommand.option(
    "--description, -o <description>",
    "Description for the tunnel",
  );

  const startTunnel = new SSHCommand(println, "start");
  startTunnel.description("Starts a tunnel");
  startTunnel.argument("<id>", "Tunnel ID to start");

  const stopTunnel = new SSHCommand(println, "stop");
  stopTunnel.description("Stops a tunnel");
  stopTunnel.argument("<id>", "Tunnel ID to stop");

  const getInbound = new SSHCommand(println, "get-inbound");
  getInbound.description("Shows all current connections");
  getInbound.argument("<id>", "Tunnel ID to view inbound connections of");
  getInbound.option("-t, --tail", "Live-view of connection list");

  const removeTunnel = new SSHCommand(println, "rm");
  removeTunnel.description("Removes a tunnel");
  removeTunnel.argument("<id>", "Tunnel ID to remove");

  program.addCommand(addCommand);
  program.addCommand(lookupCommand);
  program.addCommand(startTunnel);
  program.addCommand(stopTunnel);
  program.addCommand(getInbound);
  program.addCommand(removeTunnel);

  program.parse(argv);
}
