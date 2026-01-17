// Package narrator provides a narrative generation agent for Gas Town.
// The narrator observes work events and generates narrative content
// in configurable styles (book, tv-script, youtube-short).
package narrator

import (
	"time"

	"github.com/steveyegge/gastown/internal/agent"
)

// State is an alias for agent.State for backwards compatibility.
type State = agent.State

// State constants - re-exported from agent package for backwards compatibility.
const (
	StateStopped = agent.StateStopped
	StateRunning = agent.StateRunning
	StatePaused  = agent.StatePaused
)

// NarrativeStyle defines the output format for generated narratives.
type NarrativeStyle string

const (
	// StyleBook generates prose in the style of a novel chapter.
	StyleBook NarrativeStyle = "book"

	// StyleTVScript generates content formatted as a TV script.
	StyleTVScript NarrativeStyle = "tv-script"

	// StyleYouTubeShort generates short-form content for social media.
	StyleYouTubeShort NarrativeStyle = "youtube-short"
)

// Narrator represents the narrative generation agent.
type Narrator struct {
	// State is the current running state.
	State State `json:"state"`

	// StartedAt is when the narrator was started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// Config contains narrator configuration.
	Config NarratorConfig `json:"config"`

	// LastEventAt tracks when the last event was processed.
	LastEventAt *time.Time `json:"last_event_at,omitempty"`

	// NarrativesGenerated counts total narratives produced.
	NarrativesGenerated int `json:"narratives_generated"`
}

// NarratorConfig contains configuration for the narrator.
type NarratorConfig struct {
	// Style is the narrative output style (default: book).
	Style NarrativeStyle `json:"style"`

	// OutputDir is where generated narratives are written.
	OutputDir string `json:"output_dir,omitempty"`

	// EventTypes lists which event types to observe (empty = all).
	EventTypes []string `json:"event_types,omitempty"`

	// RigFilter limits observation to specific rigs (empty = all).
	RigFilter []string `json:"rig_filter,omitempty"`
}

// DefaultConfig returns the default narrator configuration.
func DefaultConfig() NarratorConfig {
	return NarratorConfig{
		Style: StyleBook,
	}
}
