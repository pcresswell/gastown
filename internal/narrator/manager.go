package narrator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/steveyegge/gastown/internal/agent"
	"github.com/steveyegge/gastown/internal/claude"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/session"
	"github.com/steveyegge/gastown/internal/tmux"
)

// Common errors
var (
	ErrNotRunning     = errors.New("narrator not running")
	ErrAlreadyRunning = errors.New("narrator already running")
)

// Manager handles narrator lifecycle operations.
type Manager struct {
	townRoot     string
	stateManager *agent.StateManager[Narrator]
}

// NewManager creates a new narrator manager for a town.
func NewManager(townRoot string) *Manager {
	return &Manager{
		townRoot: townRoot,
		stateManager: agent.NewStateManager[Narrator](townRoot, "narrator.json", func() *Narrator {
			return &Narrator{
				State:  StateStopped,
				Config: DefaultConfig(),
			}
		}),
	}
}

// SessionName returns the tmux session name for the narrator.
func SessionName() string {
	return session.HQPrefix + "narrator"
}

// SessionName returns the tmux session name for the narrator (method version).
func (m *Manager) SessionName() string {
	return SessionName()
}

// narratorDir returns the working directory for the narrator.
func (m *Manager) narratorDir() string {
	return filepath.Join(m.townRoot, "narrator")
}

// loadState loads narrator state from disk.
func (m *Manager) loadState() (*Narrator, error) {
	return m.stateManager.Load()
}

// saveState persists narrator state to disk using atomic write.
func (m *Manager) saveState(n *Narrator) error {
	return m.stateManager.Save(n)
}

// Start starts the narrator session.
// agentOverride allows specifying an alternate agent alias (e.g., for testing).
func (m *Manager) Start(agentOverride string) error {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	// Check if session already exists
	running, _ := t.HasSession(sessionID)
	if running {
		// Session exists - check if Claude is actually running (healthy vs zombie)
		if t.IsClaudeRunning(sessionID) {
			return ErrAlreadyRunning
		}
		// Zombie - tmux alive but Claude dead. Kill and recreate.
		if err := t.KillSession(sessionID); err != nil {
			return fmt.Errorf("killing zombie session: %w", err)
		}
	}

	// Ensure narrator directory exists
	narratorDir := m.narratorDir()
	if err := os.MkdirAll(narratorDir, 0755); err != nil {
		return fmt.Errorf("creating narrator directory: %w", err)
	}

	// Ensure Claude settings exist
	if err := claude.EnsureSettingsForRole(narratorDir, "narrator"); err != nil {
		return fmt.Errorf("ensuring Claude settings: %w", err)
	}

	// Build startup command
	startupCmd, err := config.BuildAgentStartupCommandWithAgentOverride("narrator", "", m.townRoot, "", "", agentOverride)
	if err != nil {
		return fmt.Errorf("building startup command: %w", err)
	}

	// Create session with command directly to avoid send-keys race condition.
	if err := t.NewSessionWithCommand(sessionID, narratorDir, startupCmd); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Set environment variables (non-fatal: session works without these)
	envVars := config.AgentEnv(config.AgentEnvConfig{
		Role:     "narrator",
		TownRoot: m.townRoot,
	})
	for k, v := range envVars {
		_ = t.SetEnvironment(sessionID, k, v)
	}

	// Apply narrator theming (non-fatal)
	theme := tmux.DeaconTheme() // Use deacon theme as base for now
	_ = t.ConfigureGasTownSession(sessionID, theme, "", "Narrator", "narrator")

	// Update state
	now := time.Now()
	n, _ := m.loadState()
	n.State = StateRunning
	n.StartedAt = &now
	if err := m.saveState(n); err != nil {
		_ = t.KillSession(sessionID)
		return fmt.Errorf("saving state: %w", err)
	}

	// Wait for Claude to start - fatal if Claude fails to launch
	if err := t.WaitForCommand(sessionID, constants.SupportedShells, constants.ClaudeStartTimeout); err != nil {
		_ = t.KillSessionWithProcesses(sessionID)
		return fmt.Errorf("waiting for narrator to start: %w", err)
	}

	// Accept bypass permissions warning dialog if it appears.
	_ = t.AcceptBypassPermissionsWarning(sessionID)

	time.Sleep(constants.ShutdownNotifyDelay)

	// Inject startup nudge for predecessor discovery via /resume
	_ = session.StartupNudge(t, sessionID, session.StartupNudgeConfig{
		Recipient: "narrator",
		Sender:    "daemon",
		Topic:     "observe",
	})

	// GUPP: Send propulsion nudge
	time.Sleep(2 * time.Second)
	_ = t.NudgeSession(sessionID, session.PropulsionNudgeForRole("narrator", narratorDir))

	return nil
}

// Stop stops the narrator session.
func (m *Manager) Stop() error {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	// Check if session exists
	running, err := t.HasSession(sessionID)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		return ErrNotRunning
	}

	// Try graceful shutdown first (best-effort interrupt)
	_ = t.SendKeysRaw(sessionID, "C-c")
	time.Sleep(100 * time.Millisecond)

	// Kill the session
	if err := t.KillSession(sessionID); err != nil {
		return fmt.Errorf("killing session: %w", err)
	}

	// Update state
	n, _ := m.loadState()
	n.State = StateStopped
	return m.saveState(n)
}

// IsRunning checks if the narrator session is active.
func (m *Manager) IsRunning() (bool, error) {
	t := tmux.NewTmux()
	return t.HasSession(m.SessionName())
}

// Status returns information about the narrator session.
func (m *Manager) Status() (*Narrator, error) {
	return m.loadState()
}
