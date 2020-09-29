package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"

	log "github.com/sirupsen/logrus"
)

type Endpoint struct {
	Host string
	Port int
	User string
}

func NewEndpoint(s string) *Endpoint {
	endpoint := &Endpoint{
		Host: s,
	}
	if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
		endpoint.User = parts[0]
		endpoint.Host = parts[1]
	}
	if parts := strings.Split(endpoint.Host, ":"); len(parts) > 1 {
		endpoint.Host = parts[0]
		endpoint.Port, _ = strconv.Atoi(parts[1])
	}
	return endpoint
}
func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s", endpoint.Host)
}

type Tunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
	log    *log.Logger

	Closer io.Closer
}

func (tunnel *Tunnel) logf(fmt string, args ...interface{}) {
	if tunnel.log != nil {
		fmt = "SSH Tunnel: " + fmt
		tunnel.log.Debugf(fmt, args...)
	}
}
func (tunnel *Tunnel) SetLogger(logger *log.Logger) {
	tunnel.log = logger
}

func (tunnel *Tunnel) Start() error {
	tunnel.logf("Starting")
	listener, err := net.Listen("unix", tunnel.Local.String())
	if err := os.Chmod(tunnel.Local.String(), 0700); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	tunnel.Closer = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		tunnel.logf("accepted connection")
		go tunnel.forward(conn)
	}
}

// Close closes the tunnel
func (tunnel *Tunnel) Close() {
	if tunnel.Closer == nil {
		return
	}

	tunnel.Closer.Close()
	os.Remove(tunnel.Local.String())
}

func (tunnel *Tunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", tunnel.Server.Host, tunnel.Server.Port), tunnel.Config)
	if err != nil {
		tunnel.logf("server dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())
	remoteConn, err := serverConn.Dial("unix", tunnel.Remote.String())
	if err != nil {
		tunnel.logf("remote dial error: %s", err)
		return
	}
	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())
	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("io.Copy error: %s", err)
		}
	}
	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}
func PrivateKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
func NewTunnel(tunnel string, auth []ssh.AuthMethod, localSocket string, destinationSocket string) *Tunnel {
	// A random port will be chosen for us.
	server := NewEndpoint(tunnel)
	if server.Port == 0 {
		server.Port = 22
	}
	Tunnel := &Tunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: auth,
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local:  NewEndpoint(localSocket),
		Server: server,
		Remote: NewEndpoint(destinationSocket),
	}
	return Tunnel
}
