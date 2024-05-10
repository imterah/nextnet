// @eslint-ignore-file

export type ParameterReturnedValue = {
  success: boolean;
  message?: string;
};

export type ForwardRule = {
  sourceIP: string;
  sourcePort: number;
  destPort: number;
};

export type ConnectedClient = {
  ip: string;
  port: number;

  connectionDetails: ForwardRule;
};

export class BackendBaseClass {
  state: "stopped" | "stopping" | "started" | "starting";

  clients?: ConnectedClient[]; // Not required to be implemented, but more consistency
  logs: string[];

  constructor(parameters: string) {
    this.logs = [];
    this.clients = [];

    this.state = "stopped";
  }

  addConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {}
  removeConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): void {}

  async start(): Promise<boolean> {
    return true;
  }

  async stop(): Promise<boolean> {
    return true;
  }

  getAllConnections(): ConnectedClient[] {
    if (this.clients == null) return [];
    return this.clients;
  }

  static checkParametersConnection(
    sourceIP: string,
    sourcePort: number,
    destPort: number,
    protocol: "tcp" | "udp",
  ): ParameterReturnedValue {
    return {
      success: true,
    };
  }

  static checkParametersBackendInstance(data: string): ParameterReturnedValue {
    return {
      success: true,
    };
  }
}
