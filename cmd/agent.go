package main

import (
	"cron-s/internal/scheduler"
	"github.com/judwhite/go-svc/svc"
	"github.com/spf13/cobra"
)

type agentCmd struct {
	cmd *cobra.Command
	opt *scheduler.Option
}

func newAgentCmd() *cobra.Command {
	ac := new(agentCmd)
	ac.opt = scheduler.NewOption()
	ac.cmd = &cobra.Command{
		Use:   "agent",
		Short: "Agent service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := svc.Run(scheduler.New(ac.opt)); err != nil {
				panic(err)
			}
		},
	}
	ac.addFlags()
	return ac.cmd
}

func (ac *agentCmd) addFlags() {
	ac.cmd.Flags().StringVarP(&ac.opt.HttpPort, "http-port", "", ":7570", "The HTTP API port to listen on.")
	ac.cmd.Flags().StringVarP(&ac.opt.Join, "join", "", "", "Address of another agent to join upon starting up.")
	ac.cmd.Flags().BoolVarP(&ac.opt.Raft.Bootstrap, "bootstrap", "", true, "This flag is used to control if a server is in 'bootstrap' mode.")
	ac.cmd.Flags().StringVarP(&ac.opt.Raft.NodeId, "node-id", "", "node0", "The unique ID for this server across all time.")
	ac.cmd.Flags().StringVarP(&ac.opt.Raft.Bind, "bind", "", "127.0.0.1:8570", "The address that should be bound to for internal cluster communications.")
	ac.cmd.Flags().StringVarP(&ac.opt.Raft.DataDir, "data-dir", "", "", "This flag provides a data directory for the agent to store state.")
}
