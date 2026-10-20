# Development workflow

OpenSpec is the planning and traceability layer for changes to durable `sx`
behavior. The existing Makefile, CI workflows, and technical documents remain
the authorities for executable checks and detailed product contracts.

## When to use OpenSpec

Create or select an OpenSpec change before implementing work that adds,
modifies, or removes behavior relied on by users, clients, vaults, or
automation. The proposal must name every new or modified capability so a
reviewer can see the intended contract before reading code.

Small edits do not need a made-up requirement. Refactors, tooling changes, and
documentation-only corrections that preserve durable behavior can either skip
OpenSpec or use a change that explicitly opts out of specification deltas and
explains why no capability requirements change.

## Lifecycle

```text
proposal --> specs --> design --> tasks --> apply --> verify --> archive
```

1. **Propose.** Explain the problem, scope, non-goals, affected capabilities,
   and impact.
2. **Specify.** Define observable requirements and scenarios for every affected
   capability. Do not use specs as implementation plans.
3. **Design.** Record decisions, alternatives, compatibility concerns, and
   migration where the change benefits from an explicit approach.
4. **Plan tasks.** Connect concrete repository work and its verification to the
   requirements and design. All schema-required artifacts must exist before
   implementation starts.
5. **Apply.** Work against one selected change. Mark a task complete only after
   its implementation and stated verification succeed. Leave unrelated tasks
   unchanged.
6. **Verify.** Check the implementation against the proposal, specs, design,
   and tasks. Correct the implementation or reconcile the planning artifacts
   when they differ materially.
7. **Archive.** Archive only after required tasks and verification are complete.
   Archival moves the accepted delta into the durable capability inventory.

Use the corresponding Codex OpenSpec skills for each action, or the `openspec`
CLI when working outside an integrated agent.

## Scope changes during implementation

Discovery is expected, silent scope drift is not. If implementation reveals a
material requirement, design, or scope change, update and reconcile the
planning artifacts before treating the affected task as complete. Do not make
the code fit stale artifacts or quietly narrow a requirement to finish faster.

## Specifications and reference documentation

OpenSpec owns durable capability requirements and change history. Documents
under `docs/` continue to own detailed formats, operational guidance, and
domain reference material until a focused change deliberately migrates that
capability.

When behavior described in either place changes, update the affected OpenSpec
requirement or reference document in the same change. Do not copy unchanged
reference documentation into OpenSpec: that creates two sources of truth.

## Completion and verification

Run `make prepush` after the final relevant code edit. The Makefile is the
single source of truth for what this completion gate executes.

Do not report code work complete when `make prepush` fails, cannot run, or is
interrupted. Report the unverified state and the failing check instead. When it
succeeds, summarize the result in the pull request or handoff.

Before archival, run the OpenSpec verification workflow as well. If
verification finds a material mismatch between the implementation and its
planning artifacts, keep the change active until one side is corrected.

## Pull requests

For a behavior-changing pull request, name the OpenSpec change in the pull
request template and summarize verification evidence. For work without durable
behavior impact, state that OpenSpec is not applicable and why.

Keep commit subject lines at 72 characters or fewer. Never push from an agent
session without explicit user permission.
