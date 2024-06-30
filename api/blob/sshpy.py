from typing import List, Callable
from dataclasses import dataclass
from enum import Enum

from socket import SOL_SOCKET, SO_REUSEADDR
import socketserver
import threading
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
    return ".".join(str(int.from_bytes(ip_block[i:i+1], "big")) for i in range(1, 5))
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

def convert_to_int16(arr: List[int]) -> int:
  return (arr[0] << 8) | arr[1]

def convert_int16_to_arr(num: int) -> List[int]:
  return [
    (num >> 8) & 0xff,
    num & 0xff
  ]

# Lazy, and it works.
max_tcp_sock_count = convert_to_int32([255, 255, 255, 255])

class RequestTypes(Enum):
  # Only on the server
  TCP_INITIATE_CONNECTION = 5

  # Only on the client
  TCP_INITIATE_FORWARD_RULE = 1
  UDP_INITIATE_FORWARD_RULE = 2
  TCP_CLOSE_FORWARD_RULE = 3
  UDP_CLOSE_FORWARD_RULE = 4

  # On client & server
  STATUS = 0
  
  TCP_CLOSE_CONNECTION = 6
  TCP_MESSAGE = 7
  UDP_MESSAGE = 8
  NOP = 255

class StatusTypes(Enum):
  SUCCESS = 0
  GENERAL_FAILURE = 1
  UNKNOWN_MESSAGE = 2
  MISSING_PARAMETERS = 3
  ALREADY_LISTENING  = 4

@dataclass
class TCPWrappedSocket:
  source_ip: str
  source_port: int
  dest_port: int
  has_initialized: bool

  socket: socketserver.BaseRequestHandler

class ThreadedTCPServer(socketserver.ThreadingMixIn, socketserver.TCPServer):
  pass

class ThreadedUDPServer(socketserver.ThreadingMixIn, socketserver.UDPServer):
  pass

class RequestHandler(socketserver.BaseRequestHandler):
  tcp_sockets: dict[int, TCPWrappedSocket] = {}
  tcp_current_client_id: int = 0

  tcp_servers: dict[int, ThreadedTCPServer] = {}
  udp_servers: dict[int, ThreadedUDPServer] = {}

  def read(self, byte_cnt: int) -> bytearray:
    local_buf: bytearray = bytearray(byte_cnt)
    current_bytes_read: int = 0

    while current_bytes_read != byte_cnt:
      data = self.request.recv(byte_cnt - current_bytes_read)
      
      local_buf[current_bytes_read:current_bytes_read + len(data)] = data
      current_bytes_read += len(data)

    return local_buf

  def on_tcp_callback(self, sock_server: socketserver.BaseRequestHandler):
    client_id = self.tcp_current_client_id
    client_id_calc_wraparounds = 0

    if client_id > max_tcp_sock_count:
      client_id = 0

    # Should never occur unless total clients reach the 32 bit integer limit, and then overflow
    while client_id in self.tcp_sockets:
      if client_id + 1 > max_tcp_sock_count:
        client_id = 0
        client_id_calc_wraparounds += 1
    
      if client_id_calc_wraparounds > 1:
        raise Exception("Reached protocol limit of allowed clients at once")
        
      client_id += 1
    
    self.tcp_current_client_id = client_id + 1

    client_ip, client_port = sock_server.client_address
    server_ip, server_port = sock_server.request.getsockname()

    client_wrapped_ip = make_ip_section(client_ip)
    client_wrapped_port = convert_int16_to_arr(client_port)
    server_wrapped_port = convert_int16_to_arr(server_port)

    wrapped_client_id = convert_int32_to_arr(client_id)

    tcp_socket = TCPWrappedSocket(client_ip, client_port, server_port, False, sock_server)
    self.tcp_sockets[client_id] = tcp_socket

    connection_packet = [RequestTypes.TCP_INITIATE_CONNECTION.value] + list(client_wrapped_ip) + client_wrapped_port + server_wrapped_port + wrapped_client_id
    self.request.sendall(bytes(connection_packet))

    while True:
      try:
        if tcp_socket.has_initialized:
          data = sock_server.request.recv(65535)

          if len(data) == 0:
            continue

          encoded_length = convert_int16_to_arr(len(data))
          self.request.sendall(bytes([RequestTypes.TCP_MESSAGE.value] + wrapped_client_id + encoded_length) + data)
      except (ConnectionResetError, BrokenPipeError, OSError):
        self.request.sendall(bytes([RequestTypes.TCP_CLOSE_CONNECTION.value] + wrapped_client_id))
        self.tcp_sockets.pop(client_id, None)
        return
      
  def on_udp_callback(self, sock_server: socketserver.BaseRequestHandler):
    client_ip, client_port = sock_server.client_address
    server_ip, server_port = sock_server.request[1].getsockname()

    client_wrapped_ip = make_ip_section(client_ip)
    client_wrapped_port = convert_int16_to_arr(client_port)
    server_wrapped_port = convert_int16_to_arr(server_port)

    data = sock_server.request[0]

    encoded_length = convert_int16_to_arr(len(data))
    self.request.sendall(bytes([RequestTypes.UDP_MESSAGE.value]) + client_wrapped_ip + bytes(client_wrapped_port + server_wrapped_port + encoded_length) + data)

  def handle(self):
    while True:
      original_identifier = self.read(1)

      if original_identifier[0] == RequestTypes.TCP_INITIATE_FORWARD_RULE.value:
        port_raw_byte = self.read(2)
        port = convert_to_int16(port_raw_byte)

        tcp_class = generate_forward_rule_class(self.on_tcp_callback)

        try:
          server = ThreadedTCPServer(("0.0.0.0", port), tcp_class)
          server.socket.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
          self.tcp_servers[port] = server
        except OSError:
          print(f"warn: Failed to listen on ::{port}")
          self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.GENERAL_FAILURE.value]) + original_identifier + port_raw_byte)
          continue

        print(f"info: Started listening on ::{port}")
        thread = threading.Thread(target=server.serve_forever)
        thread.daemon = True
        thread.start()

        self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.SUCCESS.value]) + original_identifier + port_raw_byte)
      elif original_identifier[0] == RequestTypes.TCP_CLOSE_FORWARD_RULE.value:
        port_raw_bytes = self.read(2)
        port = convert_to_int16(port_raw_bytes)

        if not port in self.tcp_servers:
          continue

        self.tcp_servers[port].shutdown()
        self.tcp_servers[port].socket.close()
        
        self.tcp_servers.pop(port, None)

        self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.SUCCESS.value]) + original_identifier + port_raw_byte)
      elif original_identifier[0] == RequestTypes.UDP_INITIATE_FORWARD_RULE.value:
        port_raw_byte = self.read(2)
        port = convert_to_int16(port_raw_byte)

        udp_class = generate_forward_rule_class(self.on_udp_callback)

        try:
          server = ThreadedUDPServer(("0.0.0.0", port), udp_class)
          server.socket.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)
          self.udp_servers[port] = server
        except OSError:
          print(f"warn: Failed to listen on ::{port}")
          self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.GENERAL_FAILURE.value]) + original_identifier + port_raw_byte)
          continue

        print(f"info: Started listening on ::{port}")
        thread = threading.Thread(target=server.serve_forever)
        thread.daemon = True
        thread.start()

        self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.SUCCESS.value]) + original_identifier + port_raw_byte)
      elif original_identifier[0] == RequestTypes.UDP_CLOSE_FORWARD_RULE.value:
        port_raw_bytes = self.read(2)
        port = convert_to_int16(port_raw_bytes)

        if not port in self.udp_servers:
          continue

        self.udp_servers[port].shutdown()
        self.udp_servers[port].socket.close()
        
        self.udp_servers.pop(port, None)

        self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.SUCCESS.value]) + original_identifier + port_raw_byte)
      elif original_identifier[0] == RequestTypes.TCP_MESSAGE.value:
        original_client_id = self.read(4)
        client_id = convert_to_int32(original_client_id)
        packet_len = convert_to_int16(self.read(2))

        packet = self.read(packet_len)

        if not client_id in self.tcp_sockets:
          continue

        try:
          self.tcp_sockets[client_id].socket.request.sendall(packet)
        except (ConnectionResetError, BrokenPipeError, OSError):
          self.request.sendall(bytes([RequestTypes.TCP_CLOSE_CONNECTION.value]) + original_client_id)
          self.tcp_sockets.pop(client_id, None)
      elif original_identifier[0] == RequestTypes.UDP_MESSAGE.value:
        ip_ver = self.read(1)
        ip_segment = bytearray()

        if ip_ver[0] == 4:
          ip_segment = self.read(4)
        elif ip_ver[0] == 6:
          ip_segment = self.read(16)
        else:
          print("Invalid IP segment recieved")
        
        ip_section = ip_ver + ip_segment
        ip = parse_ip_section(ip_section)

        port = convert_to_int16(self.read(2))
        
        server_port = convert_to_int16(self.read(2))
        packet_len = convert_to_int16(self.read(2))
        packet = self.read(packet_len)

        if not server_port in self.udp_servers:
          continue

        try:
          self.udp_servers[server_port].socket.sendto(packet, (ip, port))
        except (ConnectionResetError, BrokenPipeError, OSError):
          pass
      elif original_identifier[0] == RequestTypes.TCP_CLOSE_CONNECTION.value:
        client_id = convert_to_int32(self.read(4))

        if not client_id in self.tcp_sockets:
          continue

        self.tcp_sockets[client_id].socket.request.close()
        self.tcp_sockets.pop(client_id, None)
      elif original_identifier[0] == RequestTypes.NOP.value:
        pass
      elif original_identifier[0] == RequestTypes.STATUS.value:
        status_code = self.read(1)
        identifier = self.read(1)

        if status_code[0] != StatusTypes.SUCCESS.value:
          print(f"Recieved unsuccessful status code: {status_code[0]}")
          
          if identifier[0] != RequestTypes.NOP.value:
            print(f"In request type: {identifier[0]}")
        
        if identifier[0] == RequestTypes.TCP_INITIATE_CONNECTION.value:
          ip_type = self.read(1)

          # Read until we get the client ID
          if ip_type[0] == 4:
            self.read(4)
          elif ip_type[0] == 6:
            self.read(16)

          self.read(4)

          client_id = convert_to_int32(self.read(4))

          if status_code[0] == StatusTypes.SUCCESS.value and client_id in self.tcp_sockets:
            self.tcp_sockets[client_id].has_initialized = True
      else:
        self.request.sendall(bytes([RequestTypes.STATUS.value, StatusTypes.UNKNOWN_MESSAGE.value]) + original_identifier)

def generate_forward_rule_class(on_conn_callback: Callable[[socketserver.BaseRequestHandler], any]) -> socketserver.BaseRequestHandler:
  class ForwardServer(socketserver.BaseRequestHandler):
    def __init__(self, request, client_address, server):
      self.callback = on_conn_callback
      super().__init__(request, client_address, server)
    
    def handle(self):
      self.callback(self)
  
  return ForwardServer

def main():
  print("Initializing...")

  if len(sys.argv) < 2:
    print("Missing port number!")
    exit(1)

  HOST, PORT = "127.0.0.1", int(sys.argv[1])

  with ThreadedTCPServer((HOST, PORT), RequestHandler) as server:
    ip, port = server.server_address

    server.socket.setsockopt(SOL_SOCKET, SO_REUSEADDR, 1)

    print(f"Listening on {ip}:{port}")
    server.serve_forever()

if __name__ == "__main__":
  main()