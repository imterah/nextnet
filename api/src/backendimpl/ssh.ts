import { NodeSSH } from "node-ssh";
import { Socket } from "node:net";

import type { BackendBaseClass, ForwardRule, ConnectedClient, ParameterReturnedValue } from "./base.js";

type ForwardRuleExt = ForwardRule & {
  enabled: boolean
}

// Fight me (for better naming)
type BackendParsedProviderString = {
  ip:         string,
  port:       number,

  username:   string,
  privateKey: string
}

function parseBackendProviderString(data: string): BackendParsedProviderString {
  try {
    JSON.parse(data);
  } catch (e) {
    throw new Error("Payload body is not JSON")
  }

  const jsonData = JSON.parse(data);

  if (typeof jsonData.ip != "string") throw new Error("IP field is not a string");
  if (typeof jsonData.port != "number") throw new Error("Port is not a number");

  if (typeof jsonData.username != "string") throw new Error("Username is not a string");
  if (typeof jsonData.privateKey != "string") throw new Error("Private key is not a string");
  
  return {
    ip: jsonData.ip,
    port: jsonData.port,
    
    username: jsonData.username,
    privateKey: jsonData.privateKey
  }
}

export class SSHBackendProvider implements BackendBaseClass {
  state: "stopped" | "stopping" | "started" | "starting";

  clients: ConnectedClient[];
  proxies: ForwardRuleExt[];
  logs: string[];

  sshInstance: NodeSSH;
  options: BackendParsedProviderString;

  constructor(parameters: string) {
    this.logs = [];
    this.proxies = [];
    this.clients = [];
    
    this.options = parseBackendProviderString(parameters);

    this.state = "stopped";
  }

  async start(): Promise<boolean> {
    this.state = "starting";
    this.logs.push("Starting SSHBackendProvider...");

    if (this.sshInstance) {
      this.sshInstance.dispose();
    }

    this.sshInstance = new NodeSSH();

    try {
      await this.sshInstance.connect({
        host: this.options.ip,
        port: this.options.port,

        username: this.options.username,
        privateKey: this.options.privateKey
      });
    } catch (e) {
      this.logs.push(`Failed to start SSHBackendProvider! Error: '${e}'`);
      this.state = "stopped";

      // @ts-ignore
      this.sshInstance = null;

      return false;
    };

    this.state = "started";
    this.logs.push("Successfully started SSHBackendProvider.");

    return true;
  }

  async stop(): Promise<boolean> {
    this.state = "stopping";
    this.logs.push("Stopping SSHBackendProvider...");

    this.proxies.splice(0, this.proxies.length);

    this.sshInstance.dispose();

    // @ts-ignore
    this.sshInstance = null;

    this.logs.push("Successfully stopped SSHBackendProvider.");
    this.state = "stopped";

    return true;
  };

  addConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void {
    const connectionCheck = SSHBackendProvider.checkParametersConnection(sourceIP, sourcePort, destPort, protocol);
    if (!connectionCheck.success) throw new Error(connectionCheck.message);

    const foundProxyEntry = this.proxies.find((i) => i.sourceIP == sourceIP && i.sourcePort == sourcePort && i.destPort == destPort);
    if (foundProxyEntry) return;

    (async() => {
      await this.sshInstance.forwardIn("0.0.0.0", destPort, (info, accept, reject) => {
        const foundProxyEntry = this.proxies.find((i) => i.sourceIP == sourceIP && i.sourcePort == sourcePort && i.destPort == destPort);
        if (!foundProxyEntry) return reject();
        if (!foundProxyEntry.enabled) return reject();

        const client: ConnectedClient = {
          ip: info.srcIP,
          port: info.srcPort,

          connectionDetails: foundProxyEntry
        };

        this.clients.push(client);
        
        const srcConn = new Socket();
        
        srcConn.connect({
          host: sourceIP,
          port: sourcePort
        });

        // Why is this so confusing
        const destConn = accept();

        destConn.on("data", (chunk: Uint8Array) => {
          srcConn.write(chunk);
        });

        destConn.on("exit", () => {
          this.clients.splice(this.clients.indexOf(client), 1);
          srcConn.end();
        });

        srcConn.on("data", (data) => {
          destConn.write(data);
        });

        srcConn.on("end", () => {
          this.clients.splice(this.clients.indexOf(client), 1);
          destConn.close();
        });
      });
    })();

    this.proxies.push({
      sourceIP,
      sourcePort,
      destPort,

      enabled: true
    });
  };
  
  removeConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void {
    const connectionCheck = SSHBackendProvider.checkParametersConnection(sourceIP, sourcePort, destPort, protocol);
    if (!connectionCheck.success) throw new Error(connectionCheck.message);

    const foundProxyEntry = this.proxies.find((i) => i.sourceIP == sourceIP && i.sourcePort == sourcePort && i.destPort == destPort);
    if (!foundProxyEntry) return;

    foundProxyEntry.enabled = false;
  };

  getAllConnections(): ConnectedClient[] {
    return this.clients;
  };

  static checkParametersConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): ParameterReturnedValue {
    if (protocol == "udp") return {
      success: false,
      message: "SSH does not support UDP tunneling! Please use something like PortCopier instead (if it gets done)"
    };

    return {
      success: true
    }
  };

  static checkParametersBackendInstance(data: string): ParameterReturnedValue {
    try {
      parseBackendProviderString(data);
      // @ts-ignore
    } catch (e: Error) {
      return {
        success: false,
        message: e.toString()
      }
    }

    return {
      success: true
    }
  };
}