# lql-api

LQL-API is an HTTP/S and SSH bridge to the Livestatus socket, it uses the same authentication as check_mk by parsing the site’s Apache Httpasswd.

Look at [the LQL Docs](https://checkmk.com/cms_livestatus.html) to see what LQL can do for you.

LQL-API is typically installed as https://monitor.fdqn.com/“site”/lql-api/

## Commands the client supports

### localclient - Local LQL Client

```
$ lql-api localclient -h
Local LQL Client

Requires a local lql unix socket.

Examples:

- Fetch first row from the hosts table:

    $ lql-api localclient mysite -t hosts -c name -c address -c groups -l 1

- The same with stdin:

    $ echo -e "GET hosts\nColumns: name address groups\nLimit: 1" | lql-api localclient mysite

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

See the Installing docs for it.

### sshclient: SSH LQL Client

```
$ lql-api sshclient -h
SSH LQL Client

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password it will use your local agent.

Examples:

- Fetch first row from the hosts table:

    $ lql-api sshclient mysite myinternal.host.name -U mysite -t hosts -c name -c address -c groups -l 1

- The same with stdin:

    $ echo -e "GET hosts\nColumns: name address groups\nLimit: 1" | lql-api sshclient mysite myinternal.host.name -U mysite

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

Examples:

- With Debug and a single connection:

    $ lql-api sshserver mysite myinternal.host.name -d -m 1 -x 1 -U mysite

- Without Debug and maximum 5 connections:

    $ lql-api sshserver mysite myinternal.host.name -m 1 -x 5 -U mysite

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
ARGS=""
```

Now you can start the lql-api

```bash
systemctl enable --now lql-api@<site>
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

<IfModule !mod_proxy_http.c>
  Alias /<site>/lql-api/ /omd/sites/<site>
  <Directory /omd/sites/<site>/lql-api/
    Deny from all
    ErrorDocument 403 "<h1>Checkmk: Incomplete Apache Installation</h1>You need mod_proxy and
    mod_proxy_http in order to run the web interface of Checkmk."
  </Directory>
</IfModule>
```

## License

MIT - Copyright 2024 by [@jochumdev](http://github.com/jochumdev)
