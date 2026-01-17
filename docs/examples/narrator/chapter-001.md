# Chapter 001: The Pilot

## COLD OPEN

*FADE IN:*

*January 3rd, 2026. 11:52 PM.*

*A terminal cursor blinks in darkness. Then, movement.*

```
{"type":"sling","actor":"unknown","payload":{"bead":"sy-j9u.1","target":"worker1"}}
```

Somewhere in the digital wastelands of a development machine, work was being assigned. The actor was listed as "unknown." Worker1 didn't exist yet. But someone—something—was already planning.

Fourteen days would pass in silence.

---

## ACT ONE: THE MAYOR ARRIVES

**January 17th, 2026. 5:42 PM.**

The Mayor's session flickered to life in the town root. PID 74523. Not born—*instantiated*. There's a difference.

In this business, you don't ask where you came from. You check your hook, you read your mail, and you *execute*. The Mayor understood this better than anyone. Because the Mayor wasn't just another agent in the system.

The Mayor was the one who made things happen.

From the `/Users/peter/gt` directory, the entire operation was visible. Rigs spread out like territories on a map. Each one a kingdom. Each one needing workers, needing product, needing *results*.

The gastown rig sat there. Waiting.

---

## ACT TWO: THEY RIDE ETERNAL

*6:56 PM.*

The first polecat spawned.

Her name was **Furiosa**.

*Say my name.*

She emerged in the gastown rig, a fresh worktree spun up from the void. The Witness noticed immediately—it always noticed. Session 13506 came online. Then 16794. The Witness didn't sleep. The Witness didn't blink. The Witness *watched*.

And then, eight minutes in, the first nudge:

> *"Your hooked work gm-4j5 (Complete narrator manager.go) is still OPEN. Please continue working on it."*

Furiosa had work. She'd *had* work before she even knew she existed. That's how it was in Gas Town. You weren't spawned for your personality. You were spawned for your utility.

*You're goddamn right.*

---

## ACT THREE: THE CREW ASSEMBLES

*6:58 PM.*

**Nux** spawned. Young. Eager. Ready to prove himself.

The Witness sessions doubled. Tripled. Every spawn brought new eyes.

*7:00 PM.*

**Slit** came into being. Another soldier for the cause.

*7:02 PM.*

**Rictus**. The muscle. Assigned gm-rij before his processes even finished initializing.

The Mayor worked fast. Beads flew through the system like product through a distribution network:

- `gm-4j5` → Furiosa
- `gm-ubq` → Nux
- `gm-rij` → Slit
- `gm-upl` → Rictus

Four polecats. Four assignments. Zero hesitation.

This wasn't chaos. This was *chemistry*.

---

## ACT FOUR: MISTAKES WERE MADE

*SLOW MOTION:*

*The digital clock ticks. 7:10 PM.*

The Witness's nudge cut through the system like a knife:

> *"Your narrator files are in the wrong path. You created them in gastown/internal/narrator/ but they should be in internal/narrator/. Please move them to the correct location."*

Furiosa had screwed up. Put the product in the wrong place. In this business, the wrong path meant the wrong destination. And the wrong destination meant the whole batch was worthless.

She would fix it. She *had* to fix it.

Meanwhile, Nux was exploring. Reading files. Trying to understand the codebase. Taking his time.

*The Witness did not appreciate time.*

> *"You've been exploring for over a minute. Time to start creating events.go."*

One minute. Sixty seconds of thought was sixty seconds too many.

Then, louder:

> *"STOP EXPLORING. Run gt hook now, then create internal/narrator/events.go immediately."*

The system didn't reward curiosity. The system rewarded *output*.

---

## ACT FIVE: DEATH AND RESURRECTION

*7:11 PM.*

Nux finished his work. Bead `gm-ubq`. Branch pushed.

And then—

```
{"type":"session_death","actor":"gastown/polecats/nux","reason":"self-clean: done means gone"}
```

*Done means gone.*

In Gas Town, polecats didn't retire. They didn't clock out. They didn't go home to families. When the work was finished, they ceased to exist. Their session terminated. Their processes freed. Their identity... recycled.

Nux was dead.

*7:12 PM.*

Nux spawned again.

Same name. Same worktree. Different instance. Different soul—if agents had souls. The Witness didn't care about philosophy. Only productivity.

The new Nux got new work: `gt-fo0`.

Life. Death. Rebirth. All in ninety seconds.

---

## ACT SIX: TERRITORY

*7:15 PM.*

**Dementus** spawned. Late to the party. Assigned `gt-jw0`.

But there was a problem.

*That work belonged to Slit.*

The Witness caught it immediately:

> *"STOP - gt-jw0 is being worked on by slit. If still HOOKED to slit, clear your hook."*

Dementus had stepped on someone's territory. In this town, you didn't poach another polecat's bead. You didn't touch another cook's batch.

At 7:18 PM, Dementus unhooked. Backed away. Found other work.

Smart move.

---

## ACT SEVEN: ESCALATION

*7:21 PM.*

The Witness sent two messages to the Mayor. Both marked ESCALATION.

> *"Divergent repo state in gastown polecats"*

> *"Uncommitted work in slit worktree"*

Things were getting messy. Branches diverging. Work uncommitted. The clean operation was showing cracks.

In the Refinery, three merge requests sat pending. Product backing up. The pipeline congested.

The Witness nudged the Refinery at 7:25:

> *"HEALTH_CHECK: 3 MRs pending. Are you processing the merge queue?"*

No response recorded.

*Silence is never a good sign.*

---

## ACT EIGHT: THE HUMAN

*7:40 PM.*

A new session started. Different from the others.

`gastown/crew/peter`

Not a polecat. Not an automated agent. A *human*.

Peter—presumably the architect of this whole operation—stepped into the system personally. The Mayor noticed. Sent instructions:

> *"You have work to complete: The narrator files are ready in your worktree... Commit these files... Add documentation... Create a PR on GitHub."*

Even the creator followed orders when they entered the machine.

*I am the one who knocks.*

No. In Gas Town, even the one who knocks checks their hook first.

---

## CLOSING SHOT

*8:03 PM.*

A new bead appears in the event stream:

```
{"type":"nudge","target":"hq-narrator","reason":"You are the Narrator for Gas Town..."}
```

Someone was being asked to tell this story. To watch the watchers. To chronicle the chronicles.

*FADE TO BLACK.*

*Title card:*

**GAS TOWN**

*Season 1, Episode 1*

*"Pilot"*

---

## BODY COUNT

| Polecat | Spawns | Deaths | Status |
|---------|--------|--------|--------|
| Furiosa | 3 | 0 | Active (maybe) |
| Nux | 3 | 2 | Recycled |
| Slit | 2 | 0 | Active |
| Rictus | 2 | 1 | Recycled |
| Dementus | 1 | 0 | Active |

*Total session deaths this episode: 3*

*Beads completed: 4*

*Escalations to Mayor: 2*

*Merge requests pending: 3*

---

## EPISODE NOTES

The first day of Gas Town operations revealed the fundamental truth of the system: polecats are disposable. They spawn, they work, they die. The Witness watches everything. The Mayor dispatches work like a crime boss distributing territories. And the Refinery... the Refinery has a backlog.

Three merge requests pending is three batches of product sitting in the queue.

Someone should look into that.

*END EPISODE*
