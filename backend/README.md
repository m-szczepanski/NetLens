# NetLens backend

Single Go binary (`netlens`) serving the REST API, WebSocket events, and — from
Epic 2 on — the embedded SvelteKit frontend.

## Commands

```bash
make run      # go run ./cmd/netlens on 127.0.0.1:8080
make build    # compile -> bin/netlens (module is backend/, binary lands at repo root bin/)
make test     # go test ./...
make lint     # go vet + gofmt check + golangci-lint (skipped with a notice if not installed)
make fmt      # gofmt -w
make clean    # remove bin/
```

CI (`.github/workflows/ci.yml`) runs `make lint`, `go build ./...`,
`go test ./...`, plus golangci-lint with the checked-in `.golangci.yml`.

## Flags / environment

| Flag         | Env                 | Default          | Purpose                          |
| ------------ | ------------------- | ---------------- | -------------------------------- |
| `-listen`    | `NETLENS_LISTEN`    | `127.0.0.1:8080` | HTTP listen address (API, UI)    |
| `-interface` | `NETLENS_INTERFACE` | auto-detect      | interface override for discovery |

Flags override env, env overrides defaults; invalid values fail fast at startup.
Defaults are loopback-only so scan results are never exposed on the LAN by accident.

## Package boundaries

`cmd/netlens` wires `internal/config` + `internal/api`. Feature packages
(`scanner`, `discovery`, `capture`, `domain`, `fingerprint`, `storage`) are
doc.go placeholders until their epics land; each `doc.go` states the package's
responsibility and constraints. Network operations must always carry context
cancellation, timeouts, and bounded concurrency; failures must surface through
the API, not as silently incomplete scans.

## Endpoints (current contract)

```text
GET /api/status -> {"app":"netlens","version":"...","revision":"..."}
```

The frontend expects this shape (CORS headers are not sent yet — see root
README "Known gap").
