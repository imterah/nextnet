from enum import Enum

import socketserver
import threading
import signal
import sys

def sigint_handler(sig, frame):
  print("\nExiting...")
  sys.exit(0)

signal.signal(signal.SIGINT, sigint_handler) # <- no error when CTRL + C

class RequestTypes(Enum):
  TCP_INITIATE_CONNECTION = 0
  UDP_INITIATE_CONNECTION = 1
  TCP_MESSAGE = 2
  UDP_MESSAGE = 3

class ResponseTypes(Enum):
  SUCCESS = 0
  GENERAL_FAILURE = 1
  UNKNOWN_MESSAGE = 2
  MISSING_PARAMETERS = 3

class ThreadedTCPServer(socketserver.ThreadingMixIn, socketserver.TCPServer):
  pass

class RequestHandler(socketserver.BaseRequestHandler):
  def handle(self):
    print("Recieved connection")
    
    while True:
      original_identifier = self.request.recv(1)

      match original_identifier:
        case RequestTypes.TCP_INITIATE_CONNECTION:
          print("test")

        case _:
          pass

def main():
  print("Initializing...")
  HOST, PORT = "0.0.0.0", 19239

  with ThreadedTCPServer((HOST, PORT), RequestHandler) as server:
    ip, port = server.server_address

    print(f"Listening on {ip}:{port}")

    server_thread = threading.Thread(target=server.serve_forever)

    server_thread.daemon = True
    server_thread.start()
    server_thread.join()

if __name__ == "__main__":
  main()