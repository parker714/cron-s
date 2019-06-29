package main

import (
	"cron-s/internal/conf"
	"cron-s/internal/scheduler"
	"github.com/judwhite/go-svc/svc"
	"github.com/spf13/cobra"
)

type agentCmd struct {
	cmd *cobra.Command
}

func newAgentCmd() *cobra.Command {
	ac := new(agentCmd)
	ac.cmd = &cobra.Command{
		Use:   "agent",
		Short: "Agent service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Run(scheduler.New()); err != nil {
				panic(err)
			}
		},
	}

	ac.addFlags()
	return ac.cmd
}

func (ac *agentCmd) addFlags() {
	ac.cmd.Flags().StringVarP(&conf.NodeId, "node-id", "", "n0", "The unique ID for this server across all time.")
	ac.cmd.Flags().StringVarP(&conf.HttpPort, "http-port", "", ":7570", "The HTTP API port to listen on.")
	ac.cmd.Flags().BoolVarP(&conf.Bootstrap, "bootstrap", "", false, "This flag is used to control if a server is in 'bootstrap' mode.")
	ac.cmd.Flags().StringVarP(&conf.Bind, "bind", "", "127.0.0.1:8570", "The address that should be bound to for internal cluster communications.")
	ac.cmd.Flags().StringVarP(&conf.Join, "join", "", "", "Address of another agent to join upon starting up.")
	ac.cmd.Flags().StringVarP(&conf.DataDir, "data-dir", "", "data", "This flag provides a data directory for the agent to store state.")
}
