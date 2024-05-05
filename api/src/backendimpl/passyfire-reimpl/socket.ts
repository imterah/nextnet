import dgram from "node:dgram";
import net from "node:net";

import type { WebSocket } from "@fastify/websocket";
import type { FastifyRequest } from "fastify";

import type { ConnectedClientExt, PassyFireBackendProvider } from "./index.js";

// This code sucks because this protocol sucks BUUUT it works, and I don't wanna reinvent
// the gosh darn wheel for (almost) no reason

function authenticateSocket(instance: PassyFireBackendProvider, ws: WebSocket, message: string, state: ConnectedClientExt): Boolean {
  if (!message.startsWith("Accept: ")) {
    ws.send("400 Bad Request");
    return false;
  }

  const type = message.substring(message.indexOf(":") + 1).trim();

  if (type == "IsPassedWS") {
    ws.send("AcceptResponse IsPassedWS: true");
  } else if (type.startsWith("Bearer")) {
    const token = type.substring(type.indexOf("Bearer") + 7);

    for (const proxy of instance.proxies) {
      for (const username of Object.keys(proxy.userConfig)) {
        const currentToken = proxy.userConfig[username];
        
        if (token == currentToken) {
          state.connectionDetails = proxy;
          state.username = username;
        };
      };
    };

    if (state.connectionDetails && state.username) {
      ws.send("AcceptResponse Bearer: true");
      return true;
    } else {
      ws.send("AcceptResponse Bearer: false");
    }
  }

  return false;
}

export function requestHandler(instance: PassyFireBackendProvider, ws: WebSocket, req: FastifyRequest) {
  let state: "authentication" | "data" = "authentication";
  let socket: dgram.Socket | net.Socket | undefined;

  // @ts-ignore
  let connectedClient: ConnectedClientExt = {};

  ws.on("close", () => {
    instance.clients.splice(instance.clients.indexOf(connectedClient as ConnectedClientExt), 1);
  });

  ws.on("message", (rawData: ArrayBuffer) => {
    if (state == "authentication") {
      const data = rawData.toString();
      
      if (authenticateSocket(instance, ws, data, connectedClient)) {
        ws.send("AcceptResponse Bearer: true");

        connectedClient.ip = req.ip;
        connectedClient.port = req.socket.remotePort ?? -1;
        
        instance.clients.push(connectedClient);
        
        if (connectedClient.connectionDetails.protocol == "tcp") {
          socket = new net.Socket();

          socket.connect(connectedClient.connectionDetails.sourcePort, connectedClient.connectionDetails.sourceIP);

          socket.on("connect", () => {
            state = "data";

            ws.send("InitProxy: Attempting to connect");
            ws.send("InitProxy: Connected");
          });

          socket.on("data", (data) => {
            ws.send(data);
          });
        } else if (connectedClient.connectionDetails.protocol == "udp") {
          socket = dgram.createSocket("udp4");
          state = "data";

          ws.send("InitProxy: Attempting to connect");
          ws.send("InitProxy: Connected");

          socket.on("message", (data, rinfo) => {
            if (rinfo.address != connectedClient.connectionDetails.sourceIP || rinfo.port != connectedClient.connectionDetails.sourcePort) return;
            ws.send(data);
          });
        }
      }
    } else if (state == "data") {
      if (socket instanceof dgram.Socket) {
        const array = new Uint8Array(rawData);
        
        socket.send(array, connectedClient.connectionDetails.sourcePort, connectedClient.connectionDetails.sourceIP, (err) => {
          if (err) throw err;
        });
      } else if (socket instanceof net.Socket) {
        const array = new Uint8Array(rawData);

        socket.write(array);
      }
    } else {
      throw new Error(`Whooops, our WebSocket reached an unsupported state: '${state}'`);
    }
  });
}