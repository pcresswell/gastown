// Package narrator provides the narrative generation agent.
package narrator

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/steveyegge/gastown/internal/events"
)

// Significance represents the narrative importance of an event.
type Significance int

const (
	// SignificanceNone means the event should be ignored for narrative purposes.
	SignificanceNone Significance = iota
	// SignificanceLow means minor event, may be summarized or batched.
	SignificanceLow
	// SignificanceMedium means notable event, worth mentioning individually.
	SignificanceMedium
	// SignificanceHigh means major event, requires detailed narrative treatment.
	SignificanceHigh
)

// String returns a human-readable significance level.
func (s Significance) String() string {
	switch s {
	case SignificanceNone:
		return "none"
	case SignificanceLow:
		return "low"
	case SignificanceMedium:
		return "medium"
	case SignificanceHigh:
		return "high"
	default:
		return "unknown"
	}
}

// NarratorEvent wraps an events.Event with narrative metadata.
type NarratorEvent struct {
	events.Event

	// Significance is the narrative importance of this event.
	Significance Significance

	// Rig is the rig this event relates to (extracted from actor/payload).
	Rig string

	// Role is the actor's role (polecat, witness, refinery, etc).
	Role string

	// Summary is a brief human-readable description.
	Summary string
}

// EventReader reads events from .events.jsonl with offset tracking.
type EventReader struct {
	townRoot   string
	eventsPath string
	offset     int64
}

// NewEventReader creates an EventReader for the given town root.
func NewEventReader(townRoot string) *EventReader {
	return &EventReader{
		townRoot:   townRoot,
		eventsPath: filepath.Join(townRoot, events.EventsFile),
		offset:     0,
	}
}

// Offset returns the current file offset.
func (r *EventReader) Offset() int64 {
	return r.offset
}

// SetOffset sets the file offset for resuming reads.
func (r *EventReader) SetOffset(offset int64) {
	r.offset = offset
}

// ReadNew reads events added since the last read.
// Returns events and updates the internal offset.
func (r *EventReader) ReadNew() ([]NarratorEvent, error) {
	return r.readFromOffset(r.offset)
}

// ReadAll reads all events from the beginning.
// Does not update offset (use for initial load/replay).
func (r *EventReader) ReadAll() ([]NarratorEvent, error) {
	return r.readFromOffset(0)
}

// readFromOffset reads events starting from the given byte offset.
func (r *EventReader) readFromOffset(startOffset int64) ([]NarratorEvent, error) {
	file, err := os.Open(r.eventsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // No events yet
		}
		return nil, fmt.Errorf("opening events file: %w", err)
	}
	defer file.Close()

	// Seek to offset
	if startOffset > 0 {
		_, err = file.Seek(startOffset, io.SeekStart)
		if err != nil {
			return nil, fmt.Errorf("seeking to offset %d: %w", startOffset, err)
		}
	}

	var result []NarratorEvent
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		event, err := r.parseLine(line)
		if err != nil {
			// Skip malformed lines
			continue
		}

		result = append(result, event)
	}

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("reading events: %w", err)
	}

	// Update offset to current file position
	newOffset, err := file.Seek(0, io.SeekCurrent)
	if err == nil {
		r.offset = newOffset
	}

	return result, nil
}

// parseLine parses a single JSONL line into a NarratorEvent.
func (r *EventReader) parseLine(line string) (NarratorEvent, error) {
	var raw events.Event
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return NarratorEvent{}, fmt.Errorf("parsing event: %w", err)
	}

	ne := NarratorEvent{
		Event:        raw,
		Significance: classifySignificance(raw),
		Rig:          extractRig(raw),
		Role:         extractRole(raw),
		Summary:      buildSummary(raw),
	}

	return ne, nil
}

// classifySignificance determines the narrative importance of an event.
func classifySignificance(e events.Event) Significance {
	// Audit-only events are not narrative material
	if e.Visibility == events.VisibilityAudit {
		return SignificanceNone
	}

	switch e.Type {
	// High significance: major workflow events
	case events.TypeSling:
		return SignificanceHigh
	case events.TypeDone:
		return SignificanceHigh
	case events.TypeHandoff:
		return SignificanceHigh
	case events.TypeMerged:
		return SignificanceHigh
	case events.TypeMergeFailed:
		return SignificanceHigh
	case events.TypeSpawn:
		return SignificanceHigh
	case events.TypeKill:
		return SignificanceHigh
	case events.TypeBoot:
		return SignificanceHigh
	case events.TypeHalt:
		return SignificanceHigh
	case events.TypeMassDeath:
		return SignificanceHigh

	// Medium significance: notable but routine
	case events.TypeHook:
		return SignificanceMedium
	case events.TypeUnhook:
		return SignificanceMedium
	case events.TypeMail:
		return SignificanceMedium
	case events.TypeNudge:
		return SignificanceMedium
	case events.TypeSessionDeath:
		return SignificanceMedium
	case events.TypeEscalationSent:
		return SignificanceMedium
	case events.TypeEscalationAcked:
		return SignificanceMedium
	case events.TypeEscalationClosed:
		return SignificanceMedium
	case events.TypeMergeStarted:
		return SignificanceMedium
	case events.TypeMergeSkipped:
		return SignificanceMedium

	// Low significance: background activity
	case events.TypeSessionStart:
		return SignificanceLow
	case events.TypeSessionEnd:
		return SignificanceLow
	case events.TypePatrolStarted:
		return SignificanceLow
	case events.TypePatrolComplete:
		return SignificanceLow
	case events.TypePolecatChecked:
		return SignificanceLow
	case events.TypePolecatNudged:
		return SignificanceLow

	default:
		return SignificanceLow
	}
}

// extractRig extracts the rig name from an event.
func extractRig(e events.Event) string {
	// Check payload first
	if e.Payload != nil {
		if rig, ok := e.Payload["rig"].(string); ok && rig != "" {
			return rig
		}
	}

	// Extract from actor (e.g., "gastown/witness" -> "gastown")
	if e.Actor != "" {
		parts := strings.Split(e.Actor, "/")
		if len(parts) > 0 {
			first := parts[0]
			// Skip town-level actors
			if first != "mayor" && first != "deacon" && first != "gt" {
				return first
			}
		}
	}

	return ""
}

// extractRole extracts the actor's role from an event.
func extractRole(e events.Event) string {
	if e.Actor == "" {
		return ""
	}

	parts := strings.Split(e.Actor, "/")

	// Town-level roles
	if len(parts) == 1 {
		switch parts[0] {
		case "mayor", "deacon", "gt":
			return parts[0]
		}
	}

	// Rig-level roles: rig/role or rig/polecats/name
	if len(parts) >= 2 {
		second := parts[1]
		switch second {
		case "witness", "refinery", "narrator":
			return second
		case "polecats":
			return "polecat"
		case "crew":
			return "crew"
		default:
			// Might be a polecat name directly: gastown/Toast
			return "polecat"
		}
	}

	return ""
}

// buildSummary creates a brief narrative description of the event.
func buildSummary(e events.Event) string {
	switch e.Type {
	case events.TypeSling:
		bead := getPayloadString(e.Payload, "bead")
		target := getPayloadString(e.Payload, "target")
		if bead != "" && target != "" {
			return fmt.Sprintf("Work %s slung to %s", bead, target)
		}
		return "Work assignment dispatched"

	case events.TypeHook:
		bead := getPayloadString(e.Payload, "bead")
		if bead != "" {
			return fmt.Sprintf("Hooked work %s", bead)
		}
		return "Work hooked"

	case events.TypeUnhook:
		bead := getPayloadString(e.Payload, "bead")
		if bead != "" {
			return fmt.Sprintf("Unhooked work %s", bead)
		}
		return "Work unhooked"

	case events.TypeDone:
		bead := getPayloadString(e.Payload, "bead")
		if bead != "" {
			return fmt.Sprintf("Completed work %s", bead)
		}
		return "Work completed"

	case events.TypeHandoff:
		subject := getPayloadString(e.Payload, "subject")
		if subject != "" {
			return fmt.Sprintf("Handoff: %s", subject)
		}
		return "Session handoff"

	case events.TypeMail:
		to := getPayloadString(e.Payload, "to")
		subject := getPayloadString(e.Payload, "subject")
		if to != "" && subject != "" {
			return fmt.Sprintf("Mail to %s: %s", to, subject)
		}
		return "Mail sent"

	case events.TypeSpawn:
		polecat := getPayloadString(e.Payload, "polecat")
		rig := getPayloadString(e.Payload, "rig")
		if polecat != "" && rig != "" {
			return fmt.Sprintf("Spawned polecat %s in %s", polecat, rig)
		}
		return "Polecat spawned"

	case events.TypeKill:
		target := getPayloadString(e.Payload, "target")
		reason := getPayloadString(e.Payload, "reason")
		if target != "" {
			if reason != "" {
				return fmt.Sprintf("Killed %s: %s", target, reason)
			}
			return fmt.Sprintf("Killed %s", target)
		}
		return "Process killed"

	case events.TypeBoot:
		rig := getPayloadString(e.Payload, "rig")
		if rig != "" {
			return fmt.Sprintf("Booted rig %s", rig)
		}
		return "Rig booted"

	case events.TypeHalt:
		return "Services halted"

	case events.TypeNudge:
		target := getPayloadString(e.Payload, "target")
		reason := getPayloadString(e.Payload, "reason")
		if target != "" {
			if reason != "" {
				return fmt.Sprintf("Nudged %s: %s", target, reason)
			}
			return fmt.Sprintf("Nudged %s", target)
		}
		return "Agent nudged"

	case events.TypeMergeStarted:
		worker := getPayloadString(e.Payload, "worker")
		if worker != "" {
			return fmt.Sprintf("Merge started for %s", worker)
		}
		return "Merge started"

	case events.TypeMerged:
		worker := getPayloadString(e.Payload, "worker")
		if worker != "" {
			return fmt.Sprintf("Merged work from %s", worker)
		}
		return "Work merged"

	case events.TypeMergeFailed:
		worker := getPayloadString(e.Payload, "worker")
		reason := getPayloadString(e.Payload, "reason")
		if worker != "" {
			if reason != "" {
				return fmt.Sprintf("Merge failed for %s: %s", worker, reason)
			}
			return fmt.Sprintf("Merge failed for %s", worker)
		}
		return "Merge failed"

	case events.TypeMergeSkipped:
		reason := getPayloadString(e.Payload, "reason")
		if reason != "" {
			return fmt.Sprintf("Merge skipped: %s", reason)
		}
		return "Merge skipped"

	case events.TypeSessionStart:
		role := getPayloadString(e.Payload, "role")
		topic := getPayloadString(e.Payload, "topic")
		if role != "" {
			if topic != "" {
				return fmt.Sprintf("%s session started: %s", role, topic)
			}
			return fmt.Sprintf("%s session started", role)
		}
		return "Session started"

	case events.TypeSessionEnd:
		role := getPayloadString(e.Payload, "role")
		if role != "" {
			return fmt.Sprintf("%s session ended", role)
		}
		return "Session ended"

	case events.TypeSessionDeath:
		agent := getPayloadString(e.Payload, "agent")
		reason := getPayloadString(e.Payload, "reason")
		if agent != "" {
			if reason != "" {
				return fmt.Sprintf("Session died: %s (%s)", agent, reason)
			}
			return fmt.Sprintf("Session died: %s", agent)
		}
		return "Session died"

	case events.TypeMassDeath:
		count := getPayloadInt(e.Payload, "count")
		cause := getPayloadString(e.Payload, "possible_cause")
		if count > 0 {
			if cause != "" {
				return fmt.Sprintf("Mass death: %d sessions (%s)", count, cause)
			}
			return fmt.Sprintf("Mass death: %d sessions", count)
		}
		return "Mass death event"

	case events.TypePatrolStarted:
		count := getPayloadInt(e.Payload, "polecat_count")
		if count > 0 {
			return fmt.Sprintf("Patrol started (%d polecats)", count)
		}
		return "Patrol started"

	case events.TypePatrolComplete:
		count := getPayloadInt(e.Payload, "polecat_count")
		if count > 0 {
			return fmt.Sprintf("Patrol complete (%d polecats)", count)
		}
		return "Patrol complete"

	case events.TypePolecatChecked:
		polecat := getPayloadString(e.Payload, "polecat")
		status := getPayloadString(e.Payload, "status")
		if polecat != "" && status != "" {
			return fmt.Sprintf("Checked %s: %s", polecat, status)
		}
		return "Polecat checked"

	case events.TypePolecatNudged:
		polecat := getPayloadString(e.Payload, "polecat")
		if polecat != "" {
			return fmt.Sprintf("Nudged %s", polecat)
		}
		return "Polecat nudged"

	case events.TypeEscalationSent:
		target := getPayloadString(e.Payload, "target")
		to := getPayloadString(e.Payload, "to")
		if target != "" && to != "" {
			return fmt.Sprintf("Escalated %s to %s", target, to)
		}
		return "Escalation sent"

	case events.TypeEscalationAcked:
		return "Escalation acknowledged"

	case events.TypeEscalationClosed:
		return "Escalation closed"

	default:
		return e.Type
	}
}

// getPayloadString extracts a string from the event payload.
func getPayloadString(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	if v, ok := payload[key].(string); ok {
		return v
	}
	return ""
}

// getPayloadInt extracts an int from the event payload.
func getPayloadInt(payload map[string]interface{}, key string) int {
	if payload == nil {
		return 0
	}
	if v, ok := payload[key].(float64); ok {
		return int(v)
	}
	return 0
}

// FilterByRig returns events that relate to a specific rig.
func FilterByRig(evts []NarratorEvent, rig string) []NarratorEvent {
	var result []NarratorEvent
	for _, e := range evts {
		if e.Rig == rig {
			result = append(result, e)
		}
	}
	return result
}

// FilterBySignificance returns events at or above the given significance.
func FilterBySignificance(evts []NarratorEvent, minSig Significance) []NarratorEvent {
	var result []NarratorEvent
	for _, e := range evts {
		if e.Significance >= minSig {
			result = append(result, e)
		}
	}
	return result
}

// FilterByTimeRange returns events within the given time range.
func FilterByTimeRange(evts []NarratorEvent, start, end time.Time) []NarratorEvent {
	var result []NarratorEvent
	for _, e := range evts {
		ts, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			continue
		}
		if (ts.Equal(start) || ts.After(start)) && ts.Before(end) {
			result = append(result, e)
		}
	}
	return result
}

// FilterByTypes returns events matching any of the given types.
func FilterByTypes(evts []NarratorEvent, types []string) []NarratorEvent {
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	var result []NarratorEvent
	for _, e := range evts {
		if typeSet[e.Type] {
			result = append(result, e)
		}
	}
	return result
}

// ExcludeTypes returns events not matching any of the given types.
func ExcludeTypes(evts []NarratorEvent, types []string) []NarratorEvent {
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	var result []NarratorEvent
	for _, e := range evts {
		if !typeSet[e.Type] {
			result = append(result, e)
		}
	}
	return result
}
