import { resolve as pathResolver } from "node:path";
import { createSocket, type Socket as UDPSocket } from "node:dgram";
import { Socket } from "node:net";

import { NodeSSH } from "node-ssh";

import type { Channel } from "ssh2";

import type {
  BackendBaseClass,
  ForwardRule,
  ConnectedClient,
  ParameterReturnedValue,
} from "./base.js";

// For some reason the node-ssh package doesn't provide (at least, doesn't have in their docs)
// a function to copy text from one place to another. You can only use paths for some reason,
// which, to me, seems like a bad choice.
//
// So if you're confused, this is why.
const sshpyPath = pathResolver(import.meta.dirname, "../../blob/sshpy.py");

enum RequestTypes {
  // Only on the server
  TCP_INITIATE_CONNECTION = 5,

  // Only on the client
  TCP_INITIATE_FORWARD_RULE = 1,
  UDP_INITIATE_FORWARD_RULE = 2,
  TCP_CLOSE_FORWARD_RULE = 3,
  UDP_CLOSE_FORWARD_RULE = 4,

  // On client & server
  STATUS = 0,

  TCP_CLOSE_CONNECTION = 6,
  TCP_MESSAGE = 7,
  UDP_MESSAGE = 8,
  NOP = 255,
}

enum StatusTypes {
  SUCCESS = 0,
  GENERAL_FAILURE,
  UNKNOWN_MESSAGE,
  MISSING_PARAMETERS,
  ALREADY_LISTENING,
}

type BackendProvider = {
  ip: string;
  port: number;

  username: string;
  privateKey: string;

  pythonListeningPort?: number;
  pythonRuntime?: string;
};

type TCPForwardRule = {
  protocol: "tcp";
  clients: Record<number, Socket>;
};

type UDPForwardRule = {
  protocol: "udp";
  clients: VirtualPorts;
};

type ForwardRuleExt = ForwardRule & (TCPForwardRule | UDPForwardRule);

type ConnectedClientExt = {
  ip: string;
  port: number;

  connectionDetails: ForwardRuleExt;
  sock?: Socket;
};

type VirtualPortOutput = (
  message: Uint8Array | Buffer,
  ip: string,
  port: number,
) => void | Promise<void>;

class VirtualPorts {
  clients: Record<string, UDPSocket>;

  outputCallers: VirtualPortOutput[];

  ip: string;
  port: number;
  udpVer: "udp4" | "udp6";

  constructor(ip: string, port: number, udpVer?: 4 | 6) {
    this.clients = {};
    this.outputCallers = [];

    this.ip = ip;
    this.port = port;

    if (!udpVer) {
      // @ts-expect-error: This does work
      this.udpVer = "udp" + (ip.includes(":") ? 6 : 4);
    } else {
      // @ts-expect-error: This does work
      this.udpVer = "udp" + udpVer;
    }
  }

  async send(ip: string, port: number, message: Uint8Array | Buffer) {
    if (!this.clients[`${ip}:${port}`]) {
      const udpSocket = createSocket(this.udpVer);

      udpSocket.on("message", msg => {
        for (const caller of this.outputCallers) {
          try {
            caller(msg, ip, port);
          } catch (e) {
            console.error(e);
          }
        }
      });

      this.clients[`${ip}:${port}`] = udpSocket;
    }

    this.clients[`${ip}:${port}`].send(message, this.port, this.ip);
  }

  setOutput(output: VirtualPortOutput) {
    this.outputCallers.push(output);
  }
}

function makeIPSection(ip: string): Uint8Array {
  if (ip.includes(".")) {
    const ipSection = new Uint8Array(5);
    ipSection[0] = 4;

    const splitIP = ip.split(".").map(i => parseInt(i));

    // Checks if the current octet:
    // 1. Is not a number
    // 2. Is less than 0
    // 3. Is greater than 255
    const isIllegalFormat = splitIP.some(
      i => Number.isNaN(i) || i < 0 || i > 255,
    );

    // Check if it's illegal or we have too much or too little octets
    if (isIllegalFormat || splitIP.length != 4) {
      throw new Error("Illegal IPv4 address recieved");
    }

    ipSection.set(splitIP, 1);
    return ipSection;
  } else if (ip.includes(":")) {
    const ipSection = new Uint8Array(17);
    ipSection[0] = 6;

    const parsedIP: number[] = [];

    for (const splitIPSegment of ip.split(":")) {
      const splitOctetCharacters = splitIPSegment.split("");
      const octets: number[] = [];

      let octetCache: string = "";

      for (const characterIndex in splitOctetCharacters) {
        octetCache += splitOctetCharacters[characterIndex];

        // True on every other index (ex. 0 = false, 1 = true, 2 = false, etc.)
        if (parseInt(characterIndex) % 2) {
          octets.push(parseInt(octetCache, 16));
          octetCache = "";
        }
      }

      parsedIP.push(...octets);
    }

    if (parsedIP.length != 16) {
      throw new Error("Illegal IPv6 address recieved");
    }

    ipSection.set(parsedIP, 1);
    return ipSection;
  }

  throw new Error("Unknown IP format");
}

function parseIPSection(ipBlock: Uint8Array): string {
  if (ipBlock[0] == 4) {
    return ipBlock.slice(1, 5).join(".");
  } else if (ipBlock[1] == 6) {
    let address: string = "";

    const realIP = ipBlock.slice(1, 17);

    for (const octetIndexStr in realIP) {
      const octetIndex = parseInt(octetIndexStr);
      const octet = realIP[octetIndex];

      address += octet.toString(16);

      if (octetIndex % 2) {
        address += ":";
      }
    }

    return address;
  }

  throw new Error("Unknown IP format");
}

function convertToInt32(arr: Uint8Array | number[]): number {
  return (arr[0] << 24) | (arr[1] << 16) | (arr[2] << 8) | arr[3];
}

function convertInt32ToArr(num: number): Uint8Array {
  const data = new Uint8Array(4);

  data[0] = (num >> 24) & 0xff;
  data[1] = (num >> 16) & 0xff;
  data[2] = (num >> 8) & 0xff;
  data[3] = num & 0xff;

  return data;
}

function convertToInt16(arr: Uint8Array | number[]): number {
  return (arr[0] << 8) | arr[1];
}

function convertInt16ToArr(num: number): Uint8Array {
  const data = new Uint8Array(2);

  data[0] = (num >> 8) & 0xff;
  data[1] = num & 0xff;

  return data;
}

function parseBackendProviderString(data: string): BackendProvider {
  try {
    JSON.parse(data);
  } catch (e) {
    throw new Error("Payload body is not JSON");
  }

  const jsonData = JSON.parse(data);

  // SSH connection settings

  if (typeof jsonData.ip != "string") {
    throw new Error("IP field is not a string");
  }

  if (typeof jsonData.port != "number") {
    throw new Error("Port is not a number");
  }

  if (typeof jsonData.username != "string") {
    throw new Error("Username is not a string");
  }

  if (typeof jsonData.privateKey != "string") {
    throw new Error("Private key is not a string");
  }

  // SSHpy settings

  if (
    typeof jsonData.pythonListeningPort != "undefined" &&
    typeof jsonData.pythonListeningPort != "number"
  ) {
    throw new Error("Port is not a number");
  }

  if (
    typeof jsonData.pythonRuntime != "undefined" &&
    typeof jsonData.pythonRuntime != "string"
  ) {
    throw new Error("Python runtime is not a string");
  }

  return {
    ip: jsonData.ip,
    port: jsonData.port,

    username: jsonData.username,
    privateKey: jsonData.privateKey,
  };
}

export class SSHPyBackendProvider implements BackendBaseClass {
  state: "stopped" | "stopping" | "started" | "starting";

  // Proxies awaiting initialization on the server. These automatically get fully initialized in handleProtocol
  queuedProxies: ForwardRuleExt[];

  // Message buffering
  messageBuffer: Uint8Array;
  messageLastIndex: number;
  messageBufLock: boolean;

  clients: Record<number, ConnectedClientExt>;
  proxies: ForwardRuleExt[];
  logs: string[];

  sshInstance: NodeSSH;
  connection: Channel;
  options: BackendProvider;

  constructor(parameters: string) {
    this.logs = [];
    this.proxies = [];
    this.clients = {};

    this.messageBuffer = new Uint8Array(1048576 * 4);
    this.messageLastIndex = 0;
    this.messageBufLock = false;

    this.queuedProxies = [];

    this.options = parseBackendProviderString(parameters);

    this.state = "stopped";
  }

  private onStdout(chunk: Buffer) {
    const splitChunk = chunk.toString("utf8").split("\n");

    for (const chunkEntry of splitChunk) {
      if (chunkEntry == "") continue;
      this.logs.push("server: " + chunkEntry);
    }
  }

  // Someone optimize this code please.
  private async readByteCallback(): Promise<void> {
    const data: Buffer = this.connection.read();
    if (data == null) return;

    while (
      this.messageBufLock ||
      data.length + this.messageLastIndex > this.messageBuffer.length
    ) {
      await new Promise(i => setTimeout(i, 1));
    }

    this.messageBufLock = true;
    this.messageBuffer.set(data, this.messageLastIndex);
    this.messageLastIndex += data.length;
    this.messageBufLock = false;
  }

  private async readBytes(size: number): Promise<Uint8Array> {
    while (this.messageBufLock || this.messageLastIndex < size) {
      await new Promise(i => setTimeout(i, 1));
    }

    this.messageBufLock = true;
    const data = this.messageBuffer.slice(0, size);

    this.messageBuffer.set(
      this.messageBuffer.slice(size, this.messageLastIndex),
      0,
    );

    this.messageLastIndex -= size;
    this.messageBufLock = false;

    return data;
  }

  async handleProtocol() {
    while (true) {
      const messageType: number = (await this.readBytes(1))[0];

      switch (messageType) {
        default: {
          this.connection.write(
            new Uint8Array([
              RequestTypes.STATUS,
              StatusTypes.UNKNOWN_MESSAGE,
              messageType,
            ]),
          );

          break;
        }

        case RequestTypes.STATUS: {
          const statusType: number = (await this.readBytes(1))[0];
          const inResponseTo: number = (await this.readBytes(1))[0];

          if (statusType != StatusTypes.SUCCESS) {
            const statusName = StatusTypes[statusType];
            const responseName = RequestTypes[inResponseTo];

            this.logs.push(
              "ERROR: Recieved an error code from the server: " + statusName,
            );

            if (responseName != "NOP") {
              this.logs.push("Failed in: " + responseName);
            }
          }

          if (inResponseTo == RequestTypes.TCP_INITIATE_FORWARD_RULE) {
            const portRaw = await this.readBytes(2);
            const port = convertToInt16(portRaw);

            const foundProxy = this.queuedProxies.find(i => i.destPort == port);

            if (!foundProxy) {
              this.logs.push(
                "WARN: Got TCP proxy initation reply, but proxy could not be found",
              );

              this.logs.push("Please report this issue.");
              break;
            }

            if (statusType == StatusTypes.SUCCESS) {
              this.queuedProxies.splice(
                this.queuedProxies.indexOf(foundProxy),
                1,
              );

              this.proxies.push(foundProxy);
            }
          } else if (inResponseTo == RequestTypes.UDP_INITIATE_FORWARD_RULE) {
            const portRaw = await this.readBytes(2);
            const port = convertToInt16(portRaw);

            const foundProxy = this.queuedProxies.find(i => i.destPort == port);

            if (!foundProxy) {
              this.logs.push(
                "WARN: Got UDP proxy initation reply, but proxy could not be found",
              );

              this.logs.push("Please report this issue.");
              break;
            }

            if (statusType == StatusTypes.SUCCESS) {
              this.queuedProxies.splice(
                this.queuedProxies.indexOf(foundProxy),
                1,
              );

              this.proxies.push(foundProxy);
            }
          } else if (
            inResponseTo == RequestTypes.TCP_CLOSE_FORWARD_RULE ||
            inResponseTo == RequestTypes.UDP_CLOSE_FORWARD_RULE
          ) {
            const portRaw = await this.readBytes(2);
            const port = convertToInt16(portRaw);

            this.logs.push(
              "INFO: Successfully closed forward rule on port ::" + port,
            );
          }

          break;
        }

        case RequestTypes.TCP_INITIATE_CONNECTION: {
          const ipSectionType = await this.readBytes(1);
          let ipSectionData: Uint8Array;

          if (ipSectionType[0] == 4) {
            ipSectionData = await this.readBytes(4);
          } else if (ipSectionType[1] == 6) {
            ipSectionData = await this.readBytes(16);
          } else {
            break;
          }

          const ipSection = new Uint8Array(ipSectionData.length + 1);
          ipSection.set(ipSectionType, 0);
          ipSection.set(ipSectionData, 1);

          const ip = parseIPSection(ipSection);

          const clientPort = convertToInt16(await this.readBytes(2));
          const destPort = convertToInt16(await this.readBytes(2));
          const clientID = convertToInt32(await this.readBytes(4));

          const foundServer = this.proxies.find(i => i.destPort == destPort);

          if (!foundServer) {
            break;
          }

          if (foundServer.clients instanceof VirtualPorts) {
            break;
          }

          const sock = new Socket();
          foundServer.clients[clientID] = sock;

          this.clients[clientID] = {
            ip,
            port: clientPort,

            connectionDetails: foundServer,
            sock,
          };

          sock.on("data", data => {
            const tcpMessagePacket = new Uint8Array(data.length + 7);

            tcpMessagePacket[0] = RequestTypes.TCP_MESSAGE;
            tcpMessagePacket.set(convertInt32ToArr(clientID), 1);
            tcpMessagePacket.set(convertInt16ToArr(data.length), 5);
            tcpMessagePacket.set(data, 7);

            this.connection.write(tcpMessagePacket);
          });

          sock.on("close", () => {
            delete this.clients[clientID];

            if (!(foundServer.clients instanceof VirtualPorts)) {
              delete foundServer.clients[clientID];
            }

            this.connection.write(
              new Uint8Array([
                RequestTypes.TCP_CLOSE_CONNECTION,
                ...convertInt32ToArr(clientID),
              ]),
            );
          });

          sock.on("error", () => {
            try {
              sock.end();
            } catch (e) {
              //
            }

            delete this.clients[clientID];

            if (!(foundServer.clients instanceof VirtualPorts)) {
              delete foundServer.clients[clientID];
            }

            this.connection.write(
              new Uint8Array([
                RequestTypes.TCP_CLOSE_CONNECTION,
                ...convertInt32ToArr(clientID),
              ]),
            );
          });

          sock.connect(foundServer.sourcePort, foundServer.sourceIP);

          const statusMessage = new Uint8Array(ipSection.length + 11);
          statusMessage[0] = RequestTypes.STATUS;
          statusMessage[1] = StatusTypes.SUCCESS;
          statusMessage[2] = RequestTypes.TCP_INITIATE_CONNECTION;

          statusMessage.set(ipSection, 3);

          statusMessage.set(
            new Uint8Array(convertInt16ToArr(clientPort)),
            3 + ipSection.length + 0,
          );

          statusMessage.set(
            new Uint8Array(convertInt16ToArr(destPort)),
            3 + ipSection.length + 2,
          );

          statusMessage.set(
            new Uint8Array(convertInt32ToArr(clientID)),
            3 + ipSection.length + 4,
          );

          this.connection.write(statusMessage);
          break;
        }

        case RequestTypes.TCP_CLOSE_CONNECTION: {
          const clientID = convertToInt32(await this.readBytes(4));
          const client = this.clients[clientID];

          const foundServer = this.proxies.find(i =>
            i.clients instanceof VirtualPorts ? false : i.clients[clientID],
          );

          if (!client || !client.sock || !foundServer) {
            break;
          }

          client.sock.end();

          delete this.clients[clientID];

          if (!(foundServer.clients instanceof VirtualPorts)) {
            delete foundServer.clients[clientID];
          }

          break;
        }

        case RequestTypes.TCP_MESSAGE: {
          const clientID = convertToInt32(await this.readBytes(4));
          const packetLen = convertToInt16(await this.readBytes(2));
          const packet = await this.readBytes(packetLen);

          const client = this.clients[clientID];

          if (!client || !client.sock) {
            break;
          }

          client.sock.write(packet);
          break;
        }

        case RequestTypes.UDP_MESSAGE: {
          const ipSectionType = await this.readBytes(1);
          let ipSectionData: Uint8Array;

          if (ipSectionType[0] == 4) {
            ipSectionData = await this.readBytes(4);
          } else if (ipSectionType[1] == 6) {
            ipSectionData = await this.readBytes(16);
          } else {
            break;
          }

          const ipSection = new Uint8Array(ipSectionData.length + 1);
          ipSection.set(ipSectionType, 0);
          ipSection.set(ipSectionData, 1);

          const ip = parseIPSection(ipSection);

          const clientPort = convertToInt16(await this.readBytes(2));
          const destPort = convertToInt16(await this.readBytes(2));

          const packetLength = convertToInt16(await this.readBytes(2));
          const packet = await this.readBytes(packetLength);

          const proxy = this.proxies.find(i => i.destPort == destPort);

          if (!proxy || !(proxy.clients instanceof VirtualPorts)) {
            break;
          }

          proxy.clients.send(ip, clientPort, packet);

          break;
        }
      }
    }
  }

  async start(): Promise<boolean> {
    const pythonRuntime = this.options.pythonRuntime ?? "python3";
    const port = this.options.pythonListeningPort ?? 19283;

    this.state = "starting";
    this.logs.push("Starting SSHPyBackendProvider...");

    if (this.sshInstance) {
      this.sshInstance.dispose();
    }

    this.sshInstance = new NodeSSH();

    try {
      await this.sshInstance.connect({
        host: this.options.ip,
        port: this.options.port,

        username: this.options.username,
        privateKey: this.options.privateKey,
      });
    } catch (e) {
      this.logs.push(`Failed to start SSHPyBackendProvider! Error: '${e}'`);
      this.state = "stopped";

      // @ts-expect-error: We know that stuff will be initialized in order, so this will be safe
      this.sshInstance = null;
      // @ts-expect-error: Read above
      this.connection = null;

      return false;
    }

    if (this.sshInstance.connection) {
      this.sshInstance.connection.on("end", async () => {
        if (this.state != "started") return;
        this.logs.push("We disconnected from the SSH server. Restarting...");

        // Create a new array from the existing list of proxies, so we have a backup of the proxy list before
        // we wipe the list of all proxies and clients (as we're disconnected anyways)
        const proxies = Array.from(this.proxies);

        await this.stop();
        await this.start();

        if (this.state != "started") return;

        for (const proxy of proxies) {
          this.addConnection(
            proxy.sourceIP,
            proxy.sourcePort,
            proxy.destPort,
            proxy.protocol,
          );
        }
      });
    }

    this.logs.push("Successfully connected to the server. Initializing sshpy");

    // Stop & delete existing sshpy instances
    await this.sshInstance.exec("pkill", [
      "-SIGINT",
      "-f",
      `${pythonRuntime} /tmp/sshpy.py`,
    ]);

    if (
      process.env.NODE_ENV != "production" &&
      !process.env.SSHPY_BYPASS_AUTOMATIC_SERVER_SETUP
    ) {
      const serverKillPort = await this.sshInstance.exec("lsof", [
        "-t",
        `-i:${port}`,
      ]);

      if (serverKillPort) {
        await this.sshInstance.exec("kill", ["-9", serverKillPort]);
      }

      await this.sshInstance.exec("rm", ["-rf", "/tmp/sshpy.py"]);
      await this.sshInstance.putFile(sshpyPath, "/tmp/sshpy.py");

      // This is an asynchronous function, but there isn't really a good way to wait until an
      // the server starts, so we just sleep instead, and check if the port is alive.

      try {
        this.sshInstance.exec(
          pythonRuntime,
          ["-u", "/tmp/sshpy.py", `${port}`],
          {
            onStdout: this.onStdout.bind(this),
            onStderr: this.onStdout.bind(this),
          },
        );
      } catch (e) {
        this.state = "stopped";
        this.logs.push("Server (sshpy) failed to start.");

        return false;
      }
    }

    await new Promise(res => setTimeout(res, 1000));
    const serverPortIsAlive = await this.sshInstance.exec("lsof", [
      "-t",
      `-i:${port}`,
    ]);

    // If the string is empty, the server isn't running
    if (!serverPortIsAlive) {
      this.state = "stopped";
      this.logs.push("Server (sshpy) failed to start.");

      return false;
    }

    // Connect to the server running on the remote machine
    this.connection = await this.sshInstance.forwardOut(
      "127.0.0.1",
      4096,
      "127.0.0.1",
      port,
    );

    this.connection.addListener("readable", this.readByteCallback.bind(this));

    await new Promise(res => setTimeout(res, 100));

    this.handleProtocol();

    this.state = "started";
    this.logs.push("Successfully started SSHPyBackendProvider.");

    return true;
  }

  async stop(): Promise<boolean> {
    this.state = "stopping";
    this.logs.push("Stopping SSHPyBackendProvider...");

    this.queuedProxies.splice(0, this.queuedProxies.length);
    this.proxies.splice(0, this.proxies.length);
    this.clients = {};

    this.sshInstance.dispose();

    // @ts-expect-error: We know that stuff will be initialized in order, so this will be safe
    this.sshInstance = null;
    // @ts-expect-error: Read above
    this.connection = null;

    this.logs.push("Successfully stopped SSHPyBackendProvider.");
    this.state = "stopped";

    return true;
  }

  addConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    // @ts-expect-error: This should work. I don't quite know why TS is mad at me.
    const proxy: ForwardRuleExt = {
      protocol,
      sourceIP,
      sourcePort,
      destPort,

      clients: protocol == "tcp" ? [] : new VirtualPorts(sourceIP, sourcePort),
    };

    this.queuedProxies.push(proxy);

    const reqProtocol =
      protocol == "tcp"
        ? RequestTypes.TCP_INITIATE_FORWARD_RULE
        : RequestTypes.UDP_INITIATE_FORWARD_RULE;

    const reqDestPort = convertInt16ToArr(destPort);

    const connAddRequest = new Uint8Array(3);
    connAddRequest[0] = reqProtocol;
    connAddRequest.set(reqDestPort, 1);

    if (protocol == "udp" && proxy.clients instanceof VirtualPorts) {
      proxy.clients.setOutput((message, ip, port) => {
        const encodedIP = makeIPSection(ip);
        const encodedPort = convertInt16ToArr(port);

        const messageSize = convertInt16ToArr(message.length);

        const messageToSend = new Uint8Array(
          1 + 2 * 3 + encodedIP.length + message.length,
        );

        messageToSend[0] = RequestTypes.UDP_MESSAGE;

        messageToSend.set(encodedIP, 1);
        messageToSend.set(encodedPort, 1 + encodedIP.length);
        messageToSend.set(reqDestPort, 1 + encodedIP.length + 2 * 1);
        messageToSend.set(messageSize, 1 + encodedIP.length + 2 * 2);
        messageToSend.set(message, 1 + encodedIP.length + 2 * 3);

        this.connection.write(messageToSend);
      });
    }

    this.connection.write(connAddRequest);
  }

  removeConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const proxy = this.proxies.find(
      i =>
        i.sourceIP == sourceIP &&
        i.sourcePort == sourcePort &&
        i.destPort == destPort &&
        i.protocol == protocol,
    );
    if (!proxy) return;

    const reqProtocol =
      protocol == "tcp"
        ? RequestTypes.TCP_CLOSE_FORWARD_RULE
        : RequestTypes.UDP_CLOSE_FORWARD_RULE;

    const reqDestPort = convertInt16ToArr(destPort);
    const connRemoveRequest = new Uint8Array(1 + reqDestPort.length);

    connRemoveRequest[0] = reqProtocol;
    connRemoveRequest.set(reqDestPort, 1);

    this.connection.write(connRemoveRequest);

    if (protocol == "tcp") {
      const clients = proxy.clients as Record<number, Socket>;

      for (const connectionID of Object.keys(clients)) {
        clients[parseInt(connectionID)].end();
      }
    } else if (protocol == "udp" && proxy.clients instanceof VirtualPorts) {
      proxy.clients.outputCallers = [];
      proxy.clients.clients = {};
    }

    const proxyIndex = this.proxies.indexOf(proxy);
    if (proxyIndex == -1) return;

    this.proxies.splice(proxyIndex, 1);
  }

  getAllConnections(): ConnectedClient[] {
    return Object.keys(this.clients).map(i => this.clients[parseInt(i)]);
  }

  static checkParametersConnection(): ParameterReturnedValue {
    return {
      success: true,
    };
  }

  static checkParametersBackendInstance(data: string): ParameterReturnedValue {
    try {
      parseBackendProviderString(data);
      // @ts-expect-error: We write the function, and we know we're returning an error
    } catch (e: Error) {
      return {
        success: false,
        message: e.toString().replace("Error: ", ""),
      };
    }

    return {
      success: true,
    };
  }
}
