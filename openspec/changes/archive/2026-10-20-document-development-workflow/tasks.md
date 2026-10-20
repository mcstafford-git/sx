# Tasks

## 1. Configure OpenSpec project guidance

- [x] 1.1 Replace the placeholder comments in `openspec/config.yaml` with concise project context and artifact/operation guidance from the design, then verify `openspec context --json` resolves the repository without configuration errors.
- [x] 1.2 Run `openspec validate document-development-workflow --strict` after the configuration update and verify the planning artifacts still pass strict validation.

## 2. Document the contributor workflow

- [x] 2.1 Add `docs/development-workflow.md` describing the durable-behavior boundary, propose/apply/verify/archive lifecycle, relationship to existing reference docs, task tracking, and `make prepush`; verify every requirement in the `development-workflow` spec has corresponding contributor guidance.
- [x] 2.2 Add a concise link to `docs/development-workflow.md` from the README development section and verify the relative link resolves to the new document.

## 3. Add pull request traceability

- [x] 3.1 Update `.github/pull_request_template.md` with an optional OpenSpec change reference or not-applicable explanation and a place to summarize verification results; verify the existing summary, testing checklist, and notes remain present.

## 4. Verify the integrated workflow

- [x] 4.1 Review the OpenSpec change against `docs/development-workflow.md`, the README, the pull request template, and `openspec/config.yaml`; verify there are no conflicting completion rules or duplicated definitions of what `make prepush` executes.
- [x] 4.2 Run `openspec validate document-development-workflow --strict` and verify the complete change passes.
- [x] 4.3 Run `make prepush` after the final edit and record whether formatting, lint, GraphQL freshness, race-enabled tests, and build all succeed.

Verification: strict OpenSpec validation passed. `make prepush` passed formatting,
lint (0 issues), GraphQL freshness and inline-operation checks, race-enabled
tests (47 packages), and build. The first run failed because the environment's
proxy returned HTTP 502 for the Git test's intentionally invalid hostname;
the successful run added `invalid.localhost.invalid` to `NO_PROXY` and
`no_proxy` for that invocation only.

## Workflow follow-up

- Run the OpenSpec verification workflow after implementation is complete.
- Archive the change only after verification confirms the implementation matches these artifacts.
