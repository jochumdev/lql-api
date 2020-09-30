# lql-api

LQL API Client/Server for check_mk

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

## Build a debian package

if you use gvm

```bash
gvm use system; make debian; gvm use go1.15.1
```

else

```bash
make debian
```

## Installing the **localserver**

First install the package, replace "site" with your real site.

```bash
dpkg -i <package>
apt install -f
```

Next create /etc/lql-api/`site`, with the following contents:

```bash
LISTEN="localhost:8080"
DEBUG="-d"
```

Now you can start the lql-api

```bash
systemctl start lql-api@<site>
```

Next create an apache proxy for it in /etc/apache2/conf-available/zzzz_`site`_lql-api.conf

```apache
<IfModule mod_proxy_http.c>
  <Proxy http://127.0.0.1:8080/>
    Order allow,deny
    allow from all
  </Proxy>

  <Location /<site>/lql-api/>
    ProxyPass http://127.0.0.1:8080/ retry=0 timeout=120
    ProxyPassReverse http://127.0.0.1:8080/
  </Location>
</IfModule>
```

## License

MIT - Copyright 2020 by Webmeisterei GmbH
