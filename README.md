# lql_api

LQL API Server for check_mk

## Commands the client supports

### localclient - Local LQL Client

requires a local lql unix socket

### localserver: Local LQL Server

requires a local lql unix socket

### sshclient: SSH LQL Client

connects to your Server by SSH opens a SSH tunnel to the server's lql Socket and runs a query on it.

### sshserver: SSH LQL Server

Connects to your Server by SSH opens a SSH tunnel to the server's lql Socket and runs an API Server for that socket.

### Version

Prints the version

## OpenAPI 3.0 Support in sshserver

This support's OpenAPI 3.0 use the url http://localhost:8080/openapi.json and browse it over an OpenAPI browser.