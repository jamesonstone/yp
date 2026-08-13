# Testing Reference

## Purpose

- Record the project's durable commands, suites, environments, automation, and evidence expectations
- Follow `rules/testing-and-environment-validation.md` for the mandatory cross-project testing and production-safety contract
- Keep feature-specific testing details in the current feature's `SPEC.md` VALIDATION and OUTCOME sections; legacy staged flows may still use `PLAN.md` or `TASKS.md`

## Code-Level Validation

| Layer | Command | PR workflow or check | Required | Notes |
| --- | --- | --- | --- | --- |
| Unit | `go test ./...` | unavailable | yes | Exercises all packages without external services or fixtures |
| Race detection | `go test -race ./...` | unavailable | yes | Run before handoff for code changes |
| Static analysis | `go vet ./...` | unavailable | yes | Uses the Go toolchain configured by `go.mod` |
| Build | `go build ./...` | unavailable | yes | Verifies all packages compile |
| Format | `test -z "$(gofmt -l .)"` | unavailable | yes | Read-only formatting check |
| Release test | `go test ./...` | `release / goreleaser` | yes for tagged releases | Runs before GoReleaser publishes artifacts |

## High-Level Suites

| Suite | Type | Environment | Command | Automation | Evidence |
| --- | --- | --- | --- | --- | --- |
| Not applicable | none | local CLI | not required | not required | `yp` has no service boundary or deployed runtime to exercise |

## Environment Preflights

- No environment preflight is required. `yp` is a local CLI and does not connect to a deployed service.
- Interactive browsing requires `fzf`; clipboard delivery requires one of `pbcopy`, `xclip`, or `wl-copy`.

## Credentials And Test Data

- No credentials, secrets, accounts, or synthetic test data are required.
- Tests use temporary filesystem fixtures created and cleaned up by Go's test framework.

## Evidence And Retention

- Local command output and the GitHub release job are the validation evidence.
- The repository does not retain test artifacts because the test suite does not produce durable evidence files.

## Automation And Fallbacks

- Run the required code-level commands locally in the order listed above before handoff.
- The tagged-release workflow reruns `go test ./...` before publishing.

## Known Gaps

- No pull-request correctness workflow currently runs the required Go checks; they remain a local handoff responsibility.
