Golang Version SOP
==================

The targeted Go version lives in three places that must stay in lockstep. The
`tools/lint.sh` script reads the workflow value to warn when the local
toolchain drifts, so a missed file will surface as a noisy local lint run as
well as inconsistent CI behavior.

1. Update `go.mod` — set the `go` directive to the new version (e.g. `go 1.25.11`)
2. Update `.github/workflows/golang-test.yml` — set `go-version: '<version>'` under the `actions/setup-go` step
3. Update `.github/workflows/golangci-lint.yml` — set `go-version: '<version>'` under the `actions/setup-go` step
4. If the major.minor changed (e.g. 1.24 → 1.25), also check:
    - `golangci-lint` may need a bump in `.github/workflows/golangci-lint.yml` — the version pinned there must be built with a Go toolchain >= the targeted Go version, otherwise `golangci-lint run` errors with `the Go language version (goX.Y) used to build golangci-lint is lower than the targeted Go version`
    - Direct or transitive dependencies that pin a minimum Go version
5. Run `go mod tidy` to refresh `go.sum`
6. Run `go build ./...`, `go test ./...`, and `make lint` locally to confirm everything still passes
7. Note the bump in `CHANGELOG.md` under the next release's `### Changed` section if user-visible (e.g. driven by a CVE or required by a dependency); skip if purely routine maintenance
