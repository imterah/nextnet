export type ParameterReturnedValue = { 
  success: boolean,
  message?: string
}

export type ConnectedClients = {
  
};

export interface BackendInterface {
  new(): {
    addConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void;
    removeConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): void;
  
    getAllConnections(): {
      sourceIP: string,
      sourcePort: number,
      destPort: number,
      protocol: "tcp" | "udp"
    }[];

    status: "running" | "stopped" | "starting";
    logs: string[],
  },

  checkParametersConnection(sourceIP: string, sourcePort: number, destPort: number, protocol: "tcp" | "udp"): ParameterReturnedValue;
  checkParametersBackendInstance(data: string): ParameterReturnedValue;
}