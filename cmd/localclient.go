package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/webmeisterei/lql_api/lql"

	log "github.com/sirupsen/logrus"
)

func init() {
	localClientCmdLimit := 0
	localClientCmd.Flags().StringP("socket", "s", "/opt/omd/sites/{site}/tmp/run/live", "Socket on the Server")
	localClientCmd.Flags().BoolP("debug", "d", false, "Enable Debug on stderr")
	localClientCmd.Flags().StringP("format", "f", "jsonparsed", "Format one of: python, python3, json, csv, CSV, jsonparsed (default is jsonparsed, I parse json from the server)")
	localClientCmd.Flags().StringP("table", "t", "", "Produce a GET request for the given table (default: supply request by stdin)")
	localClientCmd.Flags().StringArrayP("columns", "c", []string{""}, "Columns to show from the given table, this is required if you give a table!")
	localClientCmd.Flags().StringP("user", "u", "", "LQL user to limit this on")
	localClientCmd.Flags().IntVarP(&localClientCmdLimit, "limit", "l", 0, "Limit request lines")
	rootCmd.AddCommand(localClientCmd)
}

var localClientCmd = &cobra.Command{
	Use:   "localclient [site]",
	Short: "Local LQL Client",
	Long: `Local LQL Client
	
Connects to the local socket`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sReplacer := strings.NewReplacer("{site}", args[0])
		destSocket := sReplacer.Replace(cmd.Flag("socket").Value.String())

		var lqlClient *lql.Client
		logger := log.New()
		logger.SetOutput(os.Stderr)
		if !cmd.Flag("debug").Changed {
			logger.SetLevel(log.InfoLevel)
		} else {
			logger.SetLevel(log.TraceLevel)
		}

		logger.WithField("destSocket", destSocket).Debug("Sockets")

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

			os.Exit(1)
		}(sigc)

		lqlClient, err := lql.NewClient(1, 1, "unix", destSocket)
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
