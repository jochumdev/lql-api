# lql-api

LQL API Server for check_mk

See [the LQL Docs](https://checkmk.com/cms_livestatus.html) for what LQL can do for you.

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

## OpenAPI 3.0 Support in sshserver and localserver

This support's OpenAPI 3.0 use the url http://localhost:8080/openapi.json and browse it over an OpenAPI browser.

## License

MIT - Copyright 2020 by Webmeisterei GmbH
