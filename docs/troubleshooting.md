# Troubleshooting

* I'm using SSH tunneling, and I can't reach any of the tunnels publicly.
  - Be sure to enable GatewayPorts in your sshd config (in `/etc/ssh/sshd_config` on most systems). Also, be sure to check your firewall rules on your system and your network.
