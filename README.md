# lql_api

LQL API Server for check_mk

## Commands the client supports

### localclient

Local LQL Client - requires a local lql unix socket

### sshclient

SSH LQL Client - connects to your Server by SSH opens a SSH tunnel to the server's lql Socket and runs a query on it.

### sshserver

SSH LQL Server - connects to your Server by SSH opens a SSH tunnel to the server's lql Socket and runs an API Server for that socket.

### Version

Prints the version