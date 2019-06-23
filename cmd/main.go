package main

import (
	"cron-s/internal/conf"
	"cron-s/internal/service"
	"fmt"
	"github.com/judwhite/go-svc/svc"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cron",
	Short: "Cron-s cmd",
	Run: func(cmd *cobra.Command, args []string) {
		if err := svc.Run(service.NewService()); err != nil {
			panic(err)
		}
	},
}

func main() {
	rootCmd.Flags().StringVarP(&conf.NodeId, "node-id", "", "n0", "The unique ID for this server across all time.")
	rootCmd.Flags().StringVarP(&conf.HttpPort, "http-port", "", ":7570", "The HTTP API port to listen on.")
	rootCmd.Flags().BoolVarP(&conf.Bootstrap, "bootstrap", "", false, "This flag is used to control if a server is in 'bootstrap' mode.")
	rootCmd.Flags().StringVarP(&conf.Bind, "bind", "", "127.0.0.1:8570", "The address that should be bound to for internal cluster communications.")
	rootCmd.Flags().StringVarP(&conf.Join, "join", "", "", "Address of another agent to join upon starting up.")
	rootCmd.Flags().StringVarP(&conf.DataDir, "data-dir", "", "data", "This flag provides a data directory for the agent to store state.")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
