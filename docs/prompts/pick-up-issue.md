# Pick up an issue (TDD)

A template for handing an open issue from this repo to a fresh Claude Code session
and getting a faithful, TDD-first implementation back.

## How to use

Open a new session in this repo, paste the template below into the first message,
replace `{{ISSUE_NUMBER}}` with the issue number, and (optionally) put anything
useful in `{{NOTES}}` — pointers, constraints, things you've already ruled out.
Leave everything else alone.

The template assumes:

- `CLAUDE.md`, `docs/DESIGN.md`, `docs/ARCHITECTURE.md`, `docs/ROADMAP.md`, and
  `docs/adr/README.md` are auto-loaded into the agent's context (they are — see
  the imports at the top of `CLAUDE.md`).
- `gh` is installed and authenticated for `iteratoruk/terra-firma`.
- The current workflow is small green commits straight to `main` — no per-issue
  branches yet. When that changes, update step 5.

## Template

```
Pick up issue #{{ISSUE_NUMBER}} from iteratoruk/terra-firma and implement it.

Additional notes from me: {{NOTES}}

Follow the workflow below. Project rules and design rationale are already in your
context (CLAUDE.md and the docs/ imports). The Makefile is the source of truth for
commands — read `make help`, don't infer.

1. Ground yourself.
   - `gh issue view {{ISSUE_NUMBER}}` — read the body, comments, and Gherkin in
     full. The Gherkin scenarios ARE the acceptance contract.
   - Verify the issue is open (`--json state`). If it's closed, stop and tell me.
   - Check stated dependencies. If any referenced issue is still open, stop and
     tell me before proceeding — don't try to do both.
   - Skim MEMORY.md for prior feedback that bears on this kind of work.

2. Plan in TDD slices.
   - Each Gherkin scenario maps to one (or a small group of) Go tests.
   - Order: discriminating scenarios first; conservation / property scenarios
     last (they pin what's already built).
   - If there are more than two scenarios, track them with TaskCreate so the
     state is visible.

3. The TDD loop, per scenario.
   - Write the failing test first, in the appropriate `*_test.go`. Prefer
     table-driven when the scenario naturally varies inputs.
   - Run the test and confirm it fails for the RIGHT reason (the specific
     assertion the missing behaviour cannot satisfy). A test that passes
     immediately is a test that wasn't testing anything.
   - Write the minimum code to pass. Don't anticipate later scenarios; don't
     over-build.
   - `make check` must be green before moving on.
   - Commit. Conventional-commit style, one scenario per commit where it makes
     sense. End the subject line with `(#{{ISSUE_NUMBER}})`.

4. Hold the engine invariants (CLAUDE.md) on every change.
   - New observable state surfaces through `Snapshot()` — including any
     precomputed rates (legibility is a requirement, not a nicety).
   - New mutations from outside are `Command`s. No setters.
   - The engine package imports nothing of the project (no rendering, no I/O).
   - State-affecting iteration is over canonical sorted slices, never raw maps.
   - Randomness through the world's seeded RNG. If this issue is the first to
     make the tick genuinely consume randomness, add the
     different-seed/different-world assertion currently deferred in
     `world_test.go` (`TestWorldTickIsDeterministic`, NOTE block).

5. Close out.
   - When every scenario passes and `make check` is green: `git push` (current
     workflow is straight to `main`).
   - `gh issue comment {{ISSUE_NUMBER}} --body "Implemented in <sha-range>: <one-line summary>"`.
   - `gh issue close {{ISSUE_NUMBER}}`.

If a Gherkin scenario underspecifies something — a field name, a data shape, an
edge-case behaviour — STOP and ask. Don't invent. The scenarios are the contract;
deviations need my blessing.

If a change deliberately alters the golden file: `make golden`, inspect the diff
carefully, then commit the regenerated file in the same commit as the behaviour
change. A blind regen is a bug waiting to happen.
```