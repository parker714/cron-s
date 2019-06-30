package main

import (
	"cron-s/internal/conf"
	"cron-s/internal/scheduler"
	"github.com/judwhite/go-svc/svc"
	"github.com/spf13/cobra"
)

type agentCmd struct {
	cmd *cobra.Command
	cf  *conf.Config
}

func newAgentCmd() *cobra.Command {
	ac := new(agentCmd)
	ac.cf = conf.New()
	ac.cmd = &cobra.Command{
		Use:   "agent",
		Short: "Agent service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Run(scheduler.New(ac.cf)); err != nil {
				panic(err)
			}
		},
	}
	ac.addFlags()
	return ac.cmd
}

func (ac *agentCmd) addFlags() {
	ac.cmd.Flags().StringVarP(&ac.cf.HttpPort, "http-port", "", ":7570", "The HTTP API port to listen on.")
	ac.cmd.Flags().StringVarP(&ac.cf.Join, "join", "", "", "Address of another agent to join upon starting up.")
	ac.cmd.Flags().BoolVarP(&ac.cf.Raft.Bootstrap, "bootstrap", "", false, "This flag is used to control if a server is in 'bootstrap' mode.")
	ac.cmd.Flags().StringVarP(&ac.cf.Raft.NodeId, "node-id", "", "n0", "The unique ID for this server across all time.")
	ac.cmd.Flags().StringVarP(&ac.cf.Raft.Bind, "bind", "", "127.0.0.1:8570", "The address that should be bound to for internal cluster communications.")
	ac.cmd.Flags().StringVarP(&ac.cf.Raft.DataDir, "data-dir", "", "data", "This flag provides a data directory for the agent to store state.")
}
