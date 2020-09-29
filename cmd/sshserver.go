package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	myssh "github.com/webmeisterei/lql_api/internal/ssh"
	"github.com/webmeisterei/lql_api/lql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginlogrus "github.com/toorop/gin-logrus"

	// Docs for gin swagger
	_ "github.com/webmeisterei/lql_api/docs"

	log "github.com/sirupsen/logrus"
)

func init() {
	sshServerCmd.Flags().StringP("socket", "s", "/opt/omd/sites/{site}/tmp/run/live", "Socket on the Server")
	sshServerCmd.Flags().BoolP("debug", "d", false, "Enable Debug on stderr")
	sshServerCmd.Flags().StringP("ssh-user", "U", "root", "SSH User")
	sshServerCmd.Flags().StringP("ssh-keyfile", "k", "~/.ssh/id_rsa", "Keyfile")
	sshServerCmd.Flags().StringP("ssh-password", "p", "", "Password")
	rootCmd.AddCommand(sshServerCmd)
}

var sshServerCmd = &cobra.Command{
	Use:   "sshserver [site] [server]",
	Short: "SSH LQL Server",
	Long: `SSH LQL Server

This version connects to the Check_MK Server by SSH.

If you don't provide ssh-keyfile and ssh-password I will use your local agent.
	`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sReplacer := strings.NewReplacer("{site}", args[0])
		destSocket := sReplacer.Replace(cmd.Flag("socket").Value.String())
		localSocket := path.Join(os.TempDir(), "lql-client.sock")
		var tunnel *myssh.Tunnel
		var lqlClient *lql.Client
		logger := log.New()
		logger.SetOutput(os.Stderr)
		if !cmd.Flag("debug").Changed {
			logger.SetLevel(log.InfoLevel)
		} else {
			logger.SetLevel(log.TraceLevel)
		}

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

		lqlClient, err := lql.NewClient(1, 1, "unix", localSocket)
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		defer lqlClient.Close()
		lqlClient.SetLogger(logger)

		// gin.SetMode(gin.ReleaseMode)
		r := gin.New()

		r.Use(ginlogrus.Logger(logger), gin.Recovery())

		url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

		r.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusPermanentRedirect, "/swagger/index.html")
		})

		logger.Debug("Starting the API Server")
		r.Run(":8080")
	},
}
