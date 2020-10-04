package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/webmeisterei/lql-api/lql"

	log "github.com/sirupsen/logrus"
)

func init() {
	localServerMinConns := 0
	localServerMaxConns := 0
	localServerCmd.Flags().IntVarP(&localServerMinConns, "min-conns", "m", 2, "minimal Client Connections")
	localServerCmd.Flags().IntVarP(&localServerMaxConns, "max-conns", "x", 5, "maximal Client Connections")

	localServerCmd.Flags().StringP("socket", "s", "/opt/omd/sites/{site}/tmp/run/live", "Socket")
	localServerCmd.Flags().StringP("liveproxydir", "p", "/opt/omd/sites/{site}/tmp/run/liveproxy", "Directory which contains liveproxy sockets")
	localServerCmd.Flags().StringP("htpasswd", "t", "/opt/omd/sites/{site}/etc/htpasswd", "htpasswd file")
	localServerCmd.Flags().BoolP("debug", "d", false, "Enable Debug on stderr")
	localServerCmd.Flags().StringP("listen", "l", ":8080", "Address to listen on")
	localServerCmd.Flags().StringP("logfile", "f", "/opt/omd/sites/{site}/var/log/lql-api.log", "Logfile to log to")
	rootCmd.AddCommand(localServerCmd)
}

var localServerCmd = &cobra.Command{
	Use:   "localserver [site]",
	Short: "Local LQL Server",
	Long: `Local LQL Server

Requires a local lql unix socket.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sReplacer := strings.NewReplacer("{site}", args[0])
		liveproxyDir := sReplacer.Replace(cmd.Flag("liveproxydir").Value.String())

		logfile, err := cmd.Flags().GetString("logfile")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		logfile = sReplacer.Replace(logfile)

		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()

		logger := log.New()
		logger.SetOutput(f)
		if !cmd.Flag("debug").Changed {
			logger.SetLevel(log.InfoLevel)
		} else {
			logger.SetLevel(log.TraceLevel)
		}

		socket, err := cmd.Flags().GetString("socket")
		if err != nil {
			logger.WithField("error", err).Error()
			return
		}
		localSocket := sReplacer.Replace(socket)
		var lqlClient lql.Client

		logger.WithFields(log.Fields{"localSocket": localSocket}).Debug("Sockets")

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			sig := <-c
			logger.WithFields(log.Fields{"signal": sig}).Info("Caught signal shutting down.")

			// Stop listening (and unlink the socket if unix type):
			// if lqlClient != nil {
			// 	lqlClient.Close()
			// }

			os.Exit(1)
		}(sigc)

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

		lqlClient, err = lql.NewMultiClient(minConns, maxConns, localSocket, liveproxyDir)
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
