# Architecture Decision Records

Each ADR records one significant, hard-to-reverse decision: its context, the
decision, the alternatives rejected, and the consequences. They exist so that a
settled question is not silently relitigated — especially important here, where
much of the code is agent-written and each session starts fresh. An ADR is the
durable answer to "why is it like this, and don't change it back without reading
this first."

Format: numbered `NNNN-short-title.md`. Status is one of Proposed / Accepted /
Superseded (by ADR-NNNN). Keep them short. Supersede rather than edit: an old
decision and the reason it was overturned are both valuable history.

Mechanics rationale (why the *game* works as it does) lives in DESIGN.md.
Structural maps live in ARCHITECTURE.md. ADRs are specifically for *decisions
with rejected alternatives* — the things someone might otherwise undo.

| ADR | Title | Status |
| --- | ----- | ------ |
| 0001 | Go, tick-based, not an actor model | Accepted |
