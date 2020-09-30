# lql-api

LQL API Client/Server for check_mk

See [the LQL Docs](https://checkmk.com/cms_livestatus.html) for what LQL can do for you.

## Commands the client supports

### localclient - Local LQL Client

```
$ lql-api localclient -h
Local LQL Client

Requires a local lql unix socket.

Usage:
  lql-api localclient [site] [flags]

Flags:
  -c, --columns stringArray   Columns to show from the given table, this is required if you give a table!
  -d, --debug                 Enable Debug on stderr
  -f, --format string         Format one of: python, python3, json, csv, CSV, jsonparsed (default is jsonparsed, I parse json from the server) (default "jsonparsed")
  -h, --help                  help for localclient
  -l, --limit int             Limit request lines
  -s, --socket string         Socket on the Server (default "/opt/omd/sites/{site}/tmp/run/live")
  -t, --table string          Produce a GET request for the given table (default: supply request by stdin)
  -u, --user string           LQL user to limit this on
```

### localserver: Local LQL Server

```
$ lql-api localserver -h
Local LQL Server

Requires a local lql unix socket.

Usage:
  lql-api localserver [site] [flags]

Flags:
  -d, --debug             Enable Debug on stderr
  -h, --help              help for localserver
  -t, --htpasswd string   htpasswd file (default "/opt/omd/sites/{site}/etc/htpasswd")
  -l, --listen string     Address to listen on (default ":8080")
  -x, --max-conns int     maximal Client Connections (default 5)
  -m, --min-conns int     minimal Client Connections (default 2)
  -s, --socket string     Socket (default "/opt/omd/sites/{site}/tmp/run/live")
```

### sshclient: SSH LQL Client

```
$ lql-api sshclient -h
SSH LQL Client

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password it will use your local agent.

Usage:
  lql-api sshclient [site] [server] [flags]

Flags:
  -c, --columns stringArray   Columns to show from the given table, this is required if you give a table!
  -d, --debug                 Enable Debug on stderr
  -f, --format string         Format one of: python, python3, json, csv, CSV, jsonparsed (default is jsonparsed, I parse json from the server) (default "jsonparsed")
  -h, --help                  help for sshclient
  -l, --limit int             Limit request lines
  -s, --socket string         Socket on the Server (default "/opt/omd/sites/{site}/tmp/run/live")
  -k, --ssh-keyfile string    Keyfile (default "~/.ssh/id_rsa")
  -p, --ssh-password string   Password
  -U, --ssh-user string       SSH User (default "root")
  -t, --table string          Produce a GET request for the given table (default: supply request by stdin)
  -u, --user string           LQL user to limit this on
```

### sshserver: SSH LQL Server

```
$ lql-api sshserver -h
SSH LQL Server

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password it will use your local agent.

Usage:
  lql-api sshserver [site] [server] [flags]

Flags:
  -d, --debug                 Enable Debug on stderr
  -h, --help                  help for sshserver
  -t, --htpasswd string       htpasswd file, default: NO authentication
  -l, --listen string         Address to listen on (default ":8080")
  -x, --max-conns int         maximal Client Connections (default 5)
  -m, --min-conns int         minimal Client Connections (default 2)
  -s, --socket string         Socket on the Server (default "/opt/omd/sites/{site}/tmp/run/live")
  -k, --ssh-keyfile string    Keyfile (default "~/.ssh/id_rsa")
  -p, --ssh-password string   Password
  -U, --ssh-user string       SSH User (default "root")
```

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
