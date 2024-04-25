export type ParameterReturnedValue = { 
  success: boolean,
  message?: string
}

export type ConnectedDevice = {
  sourceIP: string,
  sourcePort: number,
  destPort: number,

  protocol: "tcp" | "udp"
};

export interface BackendInterface {
  new(): {
    addConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void;
    removeConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void;
 
    run(): Promise<void>,

    getAllConnections(): {
      sourceIP: string,
      sourcePort: number,
      destPort: number,
      protocol: "tcp" | "udp"
    }[];

    state: "stopped" | "running" | "starting";

    probeConnectedClients: ConnectedDevice[],
    logs: string[]
  },

  checkParametersConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): ParameterReturnedValue;
  checkParametersBackendInstance(data: string): ParameterReturnedValue;
}