package cmd

import (
	"github.com/spf13/cobra"
)

var narratorCmd = &cobra.Command{
	Use:     "narrator",
	Aliases: []string{"nar"},
	GroupID: GroupAgents,
	Short:   "Manage the Narrator (session recorder and playback)",
	RunE:    requireSubcommand,
	Long: `Manage the Narrator - the session recorder for Gas Town.

The Narrator records and plays back agent sessions:
  - Captures agent interactions for review
  - Provides session playback capabilities
  - Enables training data generation
  - Supports debugging and audit trails

The Narrator observes without interfering with agent work.

Role shortcuts: "narrator" in mail/nudge addresses resolves to this agent.`,
}

func init() {
	rootCmd.AddCommand(narratorCmd)
}
