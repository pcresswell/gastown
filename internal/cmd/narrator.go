package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/claude"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/runtime"
	"github.com/steveyegge/gastown/internal/session"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/tmux"
	"github.com/steveyegge/gastown/internal/workspace"
)

// NarratorSessionName is the tmux session name for the Narrator.
const NarratorSessionName = "gt-narrator"

// getNarratorSessionName returns the Narrator session name.
func getNarratorSessionName() string {
	return NarratorSessionName
}

var narratorCmd = &cobra.Command{
	Use:     "narrator",
	Aliases: []string{"nar"},
	GroupID: GroupAgents,
	Short:   "Manage the Narrator (documentation and changelog agent)",
	RunE:    requireSubcommand,
	Long: `Manage the Narrator - the documentation and changelog agent for Gas Town.

The Narrator watches repository activity and maintains documentation:
  - Generates changelogs from commit history
  - Updates documentation when code changes
  - Creates release notes for version bumps
  - Summarizes large refactors for the team

The Narrator runs as a town-level service alongside Mayor and Deacon.

Role shortcuts: "narrator" in mail/nudge addresses resolves to this agent.`,
}

var narratorStartCmd = &cobra.Command{
	Use:     "start",
	Aliases: []string{"spawn"},
	Short:   "Start the Narrator session",
	Long: `Start the Narrator tmux session.

Creates a new detached tmux session for the Narrator and launches Claude.
The session runs in the workspace root directory.`,
	RunE: runNarratorStart,
}

var narratorStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Narrator session",
	Long: `Stop the Narrator tmux session.

Attempts graceful shutdown first (Ctrl-C), then kills the tmux session.`,
	RunE: runNarratorStop,
}

var narratorAttachCmd = &cobra.Command{
	Use:     "attach",
	Aliases: []string{"at"},
	Short:   "Attach to the Narrator session",
	Long: `Attach to the running Narrator tmux session.

Attaches the current terminal to the Narrator's tmux session.
Detach with Ctrl-B D.`,
	RunE: runNarratorAttach,
}

var narratorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Narrator session status",
	Long:  `Check if the Narrator tmux session is currently running.`,
	RunE:  runNarratorStatus,
}

var narratorRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Narrator session",
	Long: `Restart the Narrator tmux session.

Stops the current session (if running) and starts a fresh one.`,
	RunE: runNarratorRestart,
}

var narratorConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or modify Narrator configuration",
	Long: `Show or modify Narrator configuration.

Without arguments, displays the current configuration.
Use subcommands to modify specific settings.`,
	RunE: runNarratorConfig,
}

var narratorAgentOverride string

func init() {
	narratorCmd.AddCommand(narratorStartCmd)
	narratorCmd.AddCommand(narratorStopCmd)
	narratorCmd.AddCommand(narratorAttachCmd)
	narratorCmd.AddCommand(narratorStatusCmd)
	narratorCmd.AddCommand(narratorRestartCmd)
	narratorCmd.AddCommand(narratorConfigCmd)

	narratorStartCmd.Flags().StringVar(&narratorAgentOverride, "agent", "", "Agent alias to run the Narrator with (overrides town default)")
	narratorAttachCmd.Flags().StringVar(&narratorAgentOverride, "agent", "", "Agent alias to run the Narrator with (overrides town default)")
	narratorRestartCmd.Flags().StringVar(&narratorAgentOverride, "agent", "", "Agent alias to run the Narrator with (overrides town default)")

	rootCmd.AddCommand(narratorCmd)
}

func runNarratorStart(cmd *cobra.Command, args []string) error {
	t := tmux.NewTmux()

	sessionName := getNarratorSessionName()

	// Check if session already exists
	running, err := t.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if running {
		return fmt.Errorf("Narrator session already running. Attach with: gt narrator attach")
	}

	if err := startNarratorSession(t, sessionName, narratorAgentOverride); err != nil {
		return err
	}

	fmt.Printf("%s Narrator session started. Attach with: %s\n",
		style.Bold.Render("✓"),
		style.Dim.Render("gt narrator attach"))

	return nil
}

// startNarratorSession creates and initializes the Narrator tmux session.
func startNarratorSession(t *tmux.Tmux, sessionName, agentOverride string) error {
	// Find workspace root
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Narrator runs from its own directory (for correct role detection by gt prime)
	narratorDir := filepath.Join(townRoot, "narrator")

	// Ensure narrator directory exists
	if err := os.MkdirAll(narratorDir, 0755); err != nil {
		return fmt.Errorf("creating narrator directory: %w", err)
	}

	// Ensure Claude settings exist (autonomous role needs mail in SessionStart)
	if err := claude.EnsureSettingsForRole(narratorDir, "narrator"); err != nil {
		return fmt.Errorf("creating narrator settings: %w", err)
	}

	// Build startup command first
	// Export GT_ROLE and BD_ACTOR in the command since tmux SetEnvironment only affects new panes
	startupCmd, err := config.BuildAgentStartupCommandWithAgentOverride("narrator", "", townRoot, "", "", agentOverride)
	if err != nil {
		return fmt.Errorf("building startup command: %w", err)
	}

	// Create session with command directly to avoid send-keys race condition.
	fmt.Println("Starting Narrator session...")
	if err := t.NewSessionWithCommand(sessionName, narratorDir, startupCmd); err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	// Set environment (non-fatal: session works without these)
	// Use centralized AgentEnv for consistency across all role startup paths
	envVars := config.AgentEnv(config.AgentEnvConfig{
		Role:     "narrator",
		TownRoot: townRoot,
	})
	for k, v := range envVars {
		_ = t.SetEnvironment(sessionName, k, v)
	}

	// Apply Narrator theme (non-fatal: theming failure doesn't affect operation)
	theme := tmux.NarratorTheme()
	_ = t.ConfigureGasTownSession(sessionName, theme, "", "Narrator", "documentation")

	// Wait for Claude to start
	if err := t.WaitForCommand(sessionName, constants.SupportedShells, constants.ClaudeStartTimeout); err != nil {
		return fmt.Errorf("waiting for narrator to start: %w", err)
	}
	time.Sleep(constants.ShutdownNotifyDelay)

	runtimeConfig := config.LoadRuntimeConfig("")
	_ = runtime.RunStartupFallback(t, sessionName, "narrator", runtimeConfig)

	// Inject startup nudge for predecessor discovery via /resume
	if err := session.StartupNudge(t, sessionName, session.StartupNudgeConfig{
		Recipient: "narrator",
		Sender:    "human",
		Topic:     "documentation",
	}); err != nil {
		style.PrintWarning("failed to send startup nudge: %v", err)
	}

	// GUPP: Gas Town Universal Propulsion Principle
	// Send the propulsion nudge to trigger autonomous execution.
	// Wait for beacon to be fully processed (needs to be separate prompt)
	time.Sleep(2 * time.Second)
	if err := t.NudgeSession(sessionName, session.PropulsionNudgeForRole("narrator", narratorDir)); err != nil {
		return fmt.Errorf("sending propulsion nudge: %w", err)
	}

	return nil
}

func runNarratorStop(cmd *cobra.Command, args []string) error {
	t := tmux.NewTmux()

	sessionName := getNarratorSessionName()

	// Check if session exists
	running, err := t.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		return errors.New("Narrator session is not running")
	}

	fmt.Println("Stopping Narrator session...")

	// Try graceful shutdown first (best-effort interrupt)
	_ = t.SendKeysRaw(sessionName, "C-c")
	time.Sleep(100 * time.Millisecond)

	// Kill the session
	if err := t.KillSession(sessionName); err != nil {
		return fmt.Errorf("killing session: %w", err)
	}

	fmt.Printf("%s Narrator session stopped.\n", style.Bold.Render("✓"))
	return nil
}

func runNarratorAttach(cmd *cobra.Command, args []string) error {
	t := tmux.NewTmux()

	sessionName := getNarratorSessionName()

	// Check if session exists
	running, err := t.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		// Auto-start if not running
		fmt.Println("Narrator session not running, starting...")
		if err := startNarratorSession(t, sessionName, narratorAgentOverride); err != nil {
			return err
		}
	}

	// Use shared attach helper (smart: links if inside tmux, attaches if outside)
	return attachToTmuxSession(sessionName)
}

func runNarratorStatus(cmd *cobra.Command, args []string) error {
	t := tmux.NewTmux()

	sessionName := getNarratorSessionName()

	running, err := t.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}

	if running {
		// Get session info for more details
		info, err := t.GetSessionInfo(sessionName)
		if err == nil {
			status := "detached"
			if info.Attached {
				status = "attached"
			}
			fmt.Printf("%s Narrator session is %s\n",
				style.Bold.Render("●"),
				style.Bold.Render("running"))
			fmt.Printf("  Status: %s\n", status)
			fmt.Printf("  Created: %s\n", info.Created)
			fmt.Printf("\nAttach with: %s\n", style.Dim.Render("gt narrator attach"))
		} else {
			fmt.Printf("%s Narrator session is %s\n",
				style.Bold.Render("●"),
				style.Bold.Render("running"))
		}
	} else {
		fmt.Printf("%s Narrator session is %s\n",
			style.Dim.Render("○"),
			"not running")
		fmt.Printf("\nStart with: %s\n", style.Dim.Render("gt narrator start"))
	}

	return nil
}

func runNarratorRestart(cmd *cobra.Command, args []string) error {
	t := tmux.NewTmux()

	sessionName := getNarratorSessionName()

	running, err := t.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}

	fmt.Println("Restarting Narrator...")

	if running {
		// Kill existing session
		if err := t.KillSession(sessionName); err != nil {
			style.PrintWarning("failed to kill session: %v", err)
		}
	}

	// Start fresh
	if err := runNarratorStart(cmd, args); err != nil {
		return err
	}

	fmt.Printf("%s Narrator restarted\n", style.Bold.Render("✓"))
	fmt.Printf("  %s\n", style.Dim.Render("Use 'gt narrator attach' to connect"))
	return nil
}

func runNarratorConfig(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	narratorDir := filepath.Join(townRoot, "narrator")

	// Check if narrator directory exists
	if _, err := os.Stat(narratorDir); os.IsNotExist(err) {
		fmt.Printf("%s Narrator not configured\n", style.Dim.Render("○"))
		fmt.Printf("\nStart the Narrator to create its configuration: %s\n",
			style.Dim.Render("gt narrator start"))
		return nil
	}

	fmt.Printf("%s Narrator Configuration\n\n", style.Bold.Render("●"))
	fmt.Printf("  Directory: %s\n", narratorDir)
	fmt.Printf("  Session: %s\n", getNarratorSessionName())

	// Check for Claude settings
	settingsPath := filepath.Join(narratorDir, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Printf("  Settings: %s\n", settingsPath)
	} else {
		fmt.Printf("  Settings: %s\n", style.Dim.Render("not configured"))
	}

	return nil
}
