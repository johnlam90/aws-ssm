# Repository Guidelines

## Project Structure & Module Organization
`main.go` wires the Cobra root command, and each subcommand sits in `cmd/` (`session.go`, `eks.go`, `asg.go`, etc.). Shared logic resides in `pkg/` (`aws`, `config`, `cache`, `security`, `ui`, `testing`), documentation lives in `docs/`, runnable snippets in `examples/`, automation in `scripts/`, and release artifacts in `dist/` (clean before committing). `Formula/` and CI workflows consume this layout, so keep paths stable.

## Build, Test, and Development Commands
- `make build` – compile the local binary with version metadata; `make dev ARGS="session"` or `make run ARGS="session i-123"` run it immediately.
- `make test` – executes `go test -race -coverprofile=coverage.out ./...`; use `go test ./pkg/aws/...` to scope a package.
- `make verify` – fmt, vet, lint, and tests; must pass before PRs.
- `make lint` / `golangci-lint run ./...` – static analysis gates.
- `make test-coverage` – render `coverage.html` when you need a visual.

## Coding Style & Naming Conventions
Target Go 1.24+, rely on `go fmt` (tabs + canonical imports), and keep exported symbols documented with PascalCase names while internal helpers remain camelCase. Cobra command variables should read `<name>Cmd`. Keep CLI glue in `cmd/*`, reusable logic in `pkg/*`, and prefer small, composable functions. Run `make fmt` and `make lint` before pushing.

## Testing Guidelines
Place `_test.go` files beside the code and name tests `Test<Package>_<Behavior>`. Favor table-driven tests plus the helpers in `pkg/testing` (`TestFramework`, `Assertion`) for consistent logging. `make test` keeps race + coverage gates green, while `make test-coverage` produces the HTML report locally. Stub AWS layers or reuse fake caches instead of real credentials, and aim to leave or raise coverage before merging.

## Commit & Pull Request Guidelines
Branches follow `feature/<topic>`, `fix/<issue-id>`, or `docs/<scope>`. Commits mirror the repo history—emoji prefix plus a conventional chunk (`🔧 fix(lint): add nolint comments`), ≤50 char subject, wrapped details, and issue references (`Fixes #123`). Before opening a PR, rebase onto `upstream/main`, run `make verify`, and refresh the docs you touch. PR descriptions should summarize the change, list tests, link issues, and include screenshots or terminal snippets for user-visible updates. Focused PRs review fastest.

## Security & Configuration Tips
Review `SECURITY.md` before changing authentication or session flows. Config lives in `~/.aws-ssm/config.yaml` and cache data under `~/.aws-ssm/cache`; never hardcode secrets or sample account IDs. Respect `pkg/config` validation so new settings still honor `AWS_PROFILE`/`AWS_REGION`. If work could expose credentials, open a draft PR and contact `security+aws-ssm@johnlam.dev` instead of filing a public issue.
