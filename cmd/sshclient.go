package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	sshClientLimit := 0
	sshClientCmd.Flags().StringP("socket", "s", "/opt/omd/sites/{site}/tmp/run/live", "Socket on the Server")
	sshClientCmd.Flags().BoolP("debug", "d", false, "Enable Debug on stderr")
	sshClientCmd.Flags().StringP("ssh-user", "U", "root", "SSH User")
	sshClientCmd.Flags().StringP("ssh-keyfile", "k", "~/.ssh/id_rsa", "Keyfile")
	sshClientCmd.Flags().StringP("ssh-password", "p", "", "Password")
	sshClientCmd.Flags().StringP("format", "f", "jsonparsed", "Format one of: python, python3, json, csv, CSV, jsonparsed (default is jsonparsed, I parse json from the server)")
	sshClientCmd.Flags().StringP("table", "t", "", "Produce a GET request for the given table (default: supply request by stdin)")
	sshClientCmd.Flags().StringArrayP("columns", "c", []string{""}, "Columns to show from the given table, this is required if you give a table!")
	sshClientCmd.Flags().StringP("user", "u", "", "LQL user to limit this on")
	sshClientCmd.Flags().IntVarP(&sshClientLimit, "limit", "l", 0, "Limit request lines")
	rootCmd.AddCommand(sshClientCmd)
}

var sshClientCmd = &cobra.Command{
	Use:   "sshclient [site] [server]",
	Short: "SSH LQL Client",
	Long: `SSH LQL Client

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password it will use your local agent.

Examples:

- Fetch first row from the hosts table:

    $ lql-api sshclient mysite myinternal.host.name -U mysite -t hosts -c name -c address -c groups -l 1

- The same with stdin:

    $ echo -e "GET hosts\nColumns: name address groups\nLimit: 1" | lql-api sshclient mysite myinternal.host.name -U mysite

`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sReplacer := strings.NewReplacer("{site}", args[0])
		destSocket := sReplacer.Replace(cmd.Flag("socket").Value.String())
		localSocket := sReplacer.Replace(path.Join(os.TempDir(), "lql-{site}-client.sock"))
		var tunnel *myssh.Tunnel
		var lqlClient lql.Client
		logger := log.New()
		logger.SetOutput(os.Stderr)
		if !cmd.Flag("debug").Changed {
			logger.SetLevel(log.InfoLevel)
		} else {
			logger.SetLevel(log.TraceLevel)
		}

		logger.WithFields(log.Fields{"destSocket": destSocket, "localSocket": localSocket}).Debug("Sockets")

		var msg string
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdinBuff, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				logger.WithField("error", err).Error()
				return
			}
			logger.WithField("request", string(stdinBuff)).Debug("Got a request from stdin")
			msg = string(stdinBuff) + "\n"
		} else if cmd.Flag("table").Changed {
			if !cmd.Flag("columns").Changed {
				logger.Error("Columns is required if you want to query a table")
				return
			}
			columns, err := cmd.Flags().GetStringArray("columns")
			if err != nil {
				logger.Error("Failed to parse columns flag")
				return
			}

			msg = fmt.Sprintf("GET %s\n", cmd.Flag("table").Value.String())
			msg += fmt.Sprintf("Columns: %s\n", strings.Join(columns, " "))
		} else {
			logger.Error("Entering data interactive is not supported yet")
			return
		}

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

		lqlClient, err := lql.NewSingleClient(1, 1, "unix", localSocket)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		defer lqlClient.Close()
		lqlClient.SetLogger(logger)

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		if format != "jsonparsed" {
			result, err := lqlClient.RequestRaw(context.TODO(), msg, format, cmd.Flag("user").Value.String(), limit)
			if err != nil {
				logger.WithField("error", err).Error()
				return
			}

			os.Stdout.Write(result)
			os.Stdout.Write([]byte{'\n'})
			return
		}

		result, err := lqlClient.Request(context.TODO(), msg, cmd.Flag("user").Value.String(), limit)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		json, err := json.Marshal(result)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		os.Stdout.Write(json)
		os.Stdout.Write([]byte{'\n'})
	},
}
