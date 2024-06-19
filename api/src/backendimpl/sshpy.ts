import { resolve as pathResolver } from "node:path";
import { Socket } from "node:net";

import type { Channel } from "ssh2";
import { NodeSSH } from "node-ssh";

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

class VirtualPorts {
  constructor() {

  }
}

function makeIPSection(ip: string): Uint8Array {
  if (ip.includes(".")) {
    const ipSection = new Uint8Array(5);
    ipSection[0] = 4;

    const splitIP = ip.split(".").map((i) => parseInt(i));

    // Checks if the current octet:
    // 1. Is not a number
    // 2. Is less than 0
    // 3. Is greater than 255
    const isIllegalFormat = splitIP.some((i) => Number.isNaN(i) || i < 0 || i > 255);
    
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

enum RequestTypes {
  // Only on the server
  STATUS = 0,
  TCP_INITIATE_CONNECTION = 5,

  // Only on the client
  TCP_INITIATE_FORWARD_RULE = 1,
  UDP_INITIATE_FORWARD_RULE = 2,
  TCP_CLOSE_FORWARD_RULE = 3,
  UDP_CLOSE_FORWARD_RULE = 4,

  // On client & server
  TCP_CLOSE_CONNECTION = 6,
  TCP_MESSAGE = 7,
  UDP_MESSAGE = 8,
};

enum StatusTypes {
  SUCCESS = 0,
  GENERAL_FAILURE,
  UNKNOWN_MESSAGE,
  MISSING_PARAMETERS,
  ALREADY_LISTENING,
};

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
  clients: Socket[];
}

type UDPForwardRule = {
  protocol: "udp";
  clients: VirtualPorts;
}

type ForwardRuleExt = ForwardRule & (TCPForwardRule | UDPForwardRule);

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

  if (typeof jsonData.pythonListeningPort != "undefined" && typeof jsonData.pythonListeningPort != "number") {
    throw new Error("Port is not a number");
  }

  if (typeof jsonData.pythonRuntime != "undefined" && typeof jsonData.pythonRuntime != "string") {
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

  clients: ConnectedClient[];
  proxies: ForwardRuleExt[];
  logs: string[];

  sshInstance: NodeSSH;
  connection: Channel;
  options: BackendProvider;

  constructor(parameters: string) {
    this.logs = [];
    this.proxies = [];
    this.clients = [];

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

  private async readBytes(size: number): Promise<any> {
    let data = this.connection.read(size);

    while (data == null) {
      await new Promise((i) => setTimeout(i, 1));
      data = this.connection.read(size);
    }

    return data;
  }

  async handleProtocol() {
    while (true) {
      const data = await this.readBytes(1);
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
    await this.sshInstance.exec("pkill", ["-SIGINT", "-f", `${pythonRuntime} /tmp/sshpy.py`]);
    
    const serverKillPort = await this.sshInstance.exec("lsof", ["-t", `-i:${port}`]);
    
    if (serverKillPort) {
      await this.sshInstance.exec("kill", ["-9", serverKillPort]);
    }

    await this.sshInstance.exec("rm", ["-rf", "/tmp/sshpy.py"]);
    await this.sshInstance.putFile(sshpyPath, "/tmp/sshpy.py");
    
    // This is an asynchronous function, but there isn't really a good way to wait until an
    // the server starts, so we just sleep instead, and check if the port is alive.

    try {
      this.sshInstance.exec(pythonRuntime, ["-u", "/tmp/sshpy.py", `${port}`], {
        onStdout: this.onStdout.bind(this),
        onStderr: this.onStdout.bind(this)
      });
    } catch (e) {
      this.state = "stopped";
      this.logs.push("Server (sshpy) failed to start.");

      return false;
    }

    await new Promise((res) => setTimeout(res, 1000));
    const serverPortIsAlive = await this.sshInstance.exec("lsof", ["-t", `-i:${port}`]);
    
    // If the string is empty, the server isn't running
    if (!serverPortIsAlive) {
      this.state = "stopped";
      this.logs.push("Server (sshpy) failed to start.");

      return false;
    }

    // Connect to the server running on the remote machine
    this.connection = await this.sshInstance.forwardOut("127.0.0.1", 4096, "127.0.0.1", port);
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
    this.clients.splice(0, this.clients.length);

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
      
      clients: protocol == "tcp" ? [] : new VirtualPorts()
    };

    this.queuedProxies.push(proxy);

    const reqProtocol = protocol == "tcp" ? RequestTypes.TCP_INITIATE_FORWARD_RULE : RequestTypes.UDP_INITIATE_FORWARD_RULE;
    const reqIPSection = makeIPSection(sourceIP);
    const reqDestPort = convertInt32ToArr(destPort);

    const connAddRequest = new Uint8Array(1 + reqIPSection.length + reqDestPort.length);
    connAddRequest[0] = reqProtocol;
    connAddRequest.set(reqIPSection, 1);
    connAddRequest.set(reqDestPort, reqIPSection.length + 1);

    this.connection.write(connAddRequest);
  }

  removeConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {
    const proxy = this.proxies.find((i) => i.sourceIP == sourceIP && i.sourcePort == sourcePort && i.destPort == destPort && i.protocol == protocol);
    if (!proxy) return;

    const reqProtocol = protocol == "tcp" ? RequestTypes.TCP_CLOSE_FORWARD_RULE : RequestTypes.UDP_CLOSE_FORWARD_RULE;
    const reqIPSection = makeIPSection(sourceIP);
    const reqDestPort = convertInt32ToArr(destPort);

    const connRemoveRequest = new Uint8Array(1 + reqIPSection.length + reqDestPort.length);
    connRemoveRequest[0] = reqProtocol;
    connRemoveRequest.set(reqIPSection, 1);
    connRemoveRequest.set(reqDestPort, reqIPSection.length + 1);

    this.connection.write(connRemoveRequest);

    if (protocol == "tcp") {
      for (const connection of proxy.clients as Socket[]) {
        connection.end();
      }
    }

    const proxyIndex = this.proxies.indexOf(proxy);
    if (proxyIndex == -1) return;

    this.proxies.splice(proxyIndex, 1);
  }

  getAllConnections(): ConnectedClient[] {
    return this.clients;
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
