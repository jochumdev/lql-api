package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	myssh "github.com/webmeisterei/lql-api/internal/ssh"
	"github.com/webmeisterei/lql-api/lql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	log "github.com/sirupsen/logrus"
)

func init() {
	sshServerMinConns := 0
	sshServerMaxConns := 0
	sshServerCmd.Flags().IntVarP(&sshServerMinConns, "min-conns", "m", 2, "minimal Client Connections")
	sshServerCmd.Flags().IntVarP(&sshServerMaxConns, "max-conns", "x", 5, "maximal Client Connections")

	sshServerCmd.Flags().StringP("socket", "s", "/opt/omd/sites/{site}/tmp/run/live", "Socket on the Server")
	sshServerCmd.Flags().StringP("htpasswd", "t", "", "htpasswd file, default: NO authentication")
	sshServerCmd.Flags().BoolP("debug", "d", false, "Enable Debug on stderr")
	sshServerCmd.Flags().StringP("ssh-user", "U", "root", "SSH User")
	sshServerCmd.Flags().StringP("ssh-keyfile", "k", "~/.ssh/id_rsa", "Keyfile")
	sshServerCmd.Flags().StringP("ssh-password", "p", "", "Password")
	sshServerCmd.Flags().StringP("listen", "l", ":8080", "Address to listen on")
	rootCmd.AddCommand(sshServerCmd)
}

var sshServerCmd = &cobra.Command{
	Use:   "sshserver [site] [server]",
	Short: "SSH LQL Server",
	Long: `SSH LQL Server

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password it will use your local agent.
	`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		logger := log.New()
		logger.SetOutput(os.Stderr)
		if !cmd.Flag("debug").Changed {
			logger.SetLevel(log.InfoLevel)
		} else {
			logger.SetLevel(log.TraceLevel)
		}

		destSocket, err := cmd.Flags().GetString("socket")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		sReplacer := strings.NewReplacer("{site}", args[0])
		destSocket = sReplacer.Replace(destSocket)

		localSocket := sReplacer.Replace(path.Join(os.TempDir(), "lql-{site}-client.sock"))
		var tunnel *myssh.Tunnel
		var lqlClient *lql.Client

		logger.WithFields(log.Fields{"destSocket": destSocket, "localSocket": localSocket}).Debug("Sockets")

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			sig := <-c
			logger.WithFields(log.Fields{"signal": sig}).Info("Caught signal shutting down.")

			// Stop listening (and unlink the socket if unix type):
			if lqlClient != nil {
				lqlClient.Close()
			}
			if tunnel != nil {
				tunnel.Close()
			}

			os.Exit(1)
		}(sigc)

		if cmd.Flag("ssh-password").Changed {
			tunnel = myssh.NewTunnel(
				fmt.Sprintf("%s@%s", cmd.Flag("ssh-user").Value.String(), args[1]),
				[]ssh.AuthMethod{ssh.Password(cmd.Flag("ssh-password").Value.String())},
				localSocket,
				destSocket,
			)
		} else if cmd.Flag("ssh-password").Changed {
			tunnel = myssh.NewTunnel(
				fmt.Sprintf("%s@%s", cmd.Flag("ssh-user").Value.String(), args[1]),
				[]ssh.AuthMethod{myssh.PrivateKeyFile(cmd.Flag("ssh-keyfile").Value.String())},
				localSocket,
				destSocket,
			)
		} else {
			conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()
			ag := agent.NewClient(conn)
			auths := []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)}

			tunnel = myssh.NewTunnel(
				fmt.Sprintf("%s@%s", cmd.Flag("ssh-user").Value.String(), args[1]),
				auths,
				localSocket,
				destSocket,
			)
		}

		tunnel.SetLogger(logger)

		go tunnel.Start()
		defer tunnel.Close()
		time.Sleep(500 * time.Millisecond)

		minConns, err := cmd.Flags().GetInt("min-conns")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		maxConns, err := cmd.Flags().GetInt("max-conns")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}

		lqlClient, err = lql.NewClient(minConns, maxConns, "unix", localSocket)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		defer lqlClient.Close()
		lqlClient.SetLogger(logger)

		htpasswd := sReplacer.Replace(cmd.Flag("htpasswd").Value.String())
		server, err := lql.NewServer(lqlClient, logger, htpasswd)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}

		server.ListenAndServe(cmd.Flag("listen").Value.String())
	},
}
