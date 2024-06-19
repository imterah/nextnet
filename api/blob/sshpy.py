from enum import Enum

from typing import List
import socketserver
import signal
import sys

def sigint_handler(sig, frame):
  print("\nExiting...")
  sys.exit(0)

# Hides error & exits cleanly when SIGINT (CTRL + C, or the signal) is recieved
signal.signal(signal.SIGINT, sigint_handler)

def make_ip_section(ip: str) -> bytearray:
  if "." in ip:
    ip_section = bytearray(5)
    ip_section[0] = 4

    split_ip = [int(i) for i in ip.split(".")]
    
    # Checks if it's an impossible IPv4 address (> 0 || < 255) or we have an improper amount
    # of octets 
    if any(i < 0 or i > 255 for i in split_ip) or len(split_ip) != 4:
      raise Exception("Illegal IPv4 IP address recieved")

    ip_section[1:] = split_ip
    return ip_section
  elif ":" in ip:
    ip_section = bytearray(17)
    ip_section[0] = 6

    parsed_ip = []

    for split_ip_segment in ip.split(":"):
      split_octet_characters = split_ip_segment.split("")
      octets: List[int] = []

      octet_cache = ""

      for character_index in range(len(split_octet_characters)):
        octet_cache += split_octet_characters[character_index]

        if character_index % 2:
          octets.append(int(octet_cache, 16))
          octet_cache = ""
    
      parsed_ip.extend(octets)
    
    if len(parsed_ip) != 16:
      raise Exception("Illegal IPv6 address recieved")
    
    ip_section[1:] = parsed_ip
    return ip_section
  
  raise Exception("Unknown IP format")

def parse_ip_section(ip_block: bytearray) -> str:
  if ip_block[0] == 4:
    return ".".join(str(int.from_bytes(ip_block[i:i+1])) for i in range(1, 5))
  elif ip_block[1] == 6:
    address = ""

    real_ip = ip_block[1:17]

    for octet_index in len(real_ip):
      octet = real_ip[octet_index]
      address += octet.hex()

      if octet_index % 2:
        address += ":"
    
    return address
  
  raise Exception("Unknown IP format")

def convert_to_int32(arr: List[int]) -> int:
  return (arr[0] << 24) | (arr[1] << 16) | (arr[2] << 8) | arr[3]

def convert_int32_to_arr(num: int) -> List[int]:
  return [
    (num >> 24) & 0xff,
    (num >> 16) & 0xff,
    (num >> 8) & 0xff,
    num & 0xff
  ]

class RequestTypes(Enum):
  # Only on the server
  STATUS = 0
  TCP_INITIATE_CONNECTION = 3

  # Only on the client
  TCP_INITIATE_FORWARD_RULE = 1
  UDP_INITIATE_FORWARD_RULE = 2
  TCP_CLOSE_FORWARD_RULE = 3
  UDP_CLOSE_FORWARD_RULE = 4

  # On client & server
  TCP_CLOSE_CONNECTION = 4
  TCP_MESSAGE = 5
  UDP_MESSAGE = 6

class StatusTypes(Enum):
  SUCCESS = 0
  GENERAL_FAILURE = 1
  UNKNOWN_MESSAGE = 2
  MISSING_PARAMETERS = 3
  ALREADY_LISTENING  = 4

class ThreadedTCPServer(socketserver.ThreadingMixIn, socketserver.TCPServer):
  pass

class RequestHandler(socketserver.BaseRequestHandler):
  message_queue = []
  
  def handle(self):    
    while True:
      for message in self.message_queue:
        self.request.sendall(message)
    
      original_identifier = self.request.recv(1)

      match original_identifier:
        case _:
          self.request.sendall([RequestTypes.STATUS, StatusTypes.UNKNOWN_MESSAGE])
          pass

def main():
  print("Initializing...")

  if len(sys.argv) < 2:
    print("Missing port number!")
    exit(1)

  HOST, PORT = "127.0.0.1", int(sys.argv[1])

  with ThreadedTCPServer((HOST, PORT), RequestHandler) as server:
    ip, port = server.server_address

    print(f"Listening on {ip}:{port}")
    server.serve_forever()

if __name__ == "__main__":
  main()