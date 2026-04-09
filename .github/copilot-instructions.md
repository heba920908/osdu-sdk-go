# OSDU Go SDK — Project Guidelines

## Architecture

The SDK has two client generations in `pkg/osdu/`:

- **v1** (`pkg/osdu/`): `OsduApiRequest` struct, constructor variants `NewClient()` / `NewClientWithProvider()` / `NewClientWithConfig()`
- **v2** (`pkg/osdu/v2/`): `OsduClient` interface with functional options (`WithAuthProvider`, `WithOsduSettings`, `WithHTTPClient`, `WithContext`, `WithLogger`)

Auth is fully pluggable via the `auth.AuthProvider` interface (`pkg/auth/interface.go`). The only built-in provider is OpenID Connect (`pkg/auth/openid.go`). Custom providers (e.g., Azure) live in a **separate module** (`examples/auth_azure_mod_example/`) to avoid pulling in the heavy Azure SDK as a mandatory dependency.

Key packages:
- `pkg/auth/` — `AuthProvider` interface, `ProviderFactory`, built-in OpenID provider
- `pkg/config/` — YAML config loading (`CONFIG_FILE` env var, defaults to `./config/default.yaml`), env-override helpers
- `pkg/models/` — Domain types (partitions, workflows, datasets, files, etc.)
- `pkg/utils/` — Webhook signature verification

## Build & Test

```bash
go mod download
go build -v ./...
go test -v ./pkg/osdu
go test -v ./pkg/auth

# Run a specific test
go test -v ./pkg/osdu/ -run TestPartitionProvisioning
```

Required env vars for integration tests:
```bash
export OSDU_AUTH_CLIENT_ID=datafier
export OSDU_AUTH_CLIENT_SECRET=<secret>
```

The Azure example is a **separate Go module** — work inside `examples/auth_azure_mod_example/` and use its own `go.mod`.

## Conventions

**Testing**: Use `testify/mock` for `AuthProvider` mocks, `testify/assert` for assertions. Tests inject a mock via `NewClientWithConfig()`. See `pkg/osdu/client_test.go` for the `createMockClient()` helper pattern.

**Logging**: Use `log/slog` with context-aware calls (`slog.InfoContext(ctx, msg, "key", val)`). Do not use `fmt.Println` or `log.Print*`.

**Error handling**: Return `(value, error)`. Wrap errors with context: `fmt.Errorf("service response - %d: %s", res.StatusCode, body)`. HTTP status `> 205` is an error. Status `409` (conflict/already exists) is handled gracefully in several services.

**New auth providers**: Implement `auth.AuthProvider`; if the provider requires external dependencies, add it as a separate module in `examples/` following the Azure pattern. Document in `docs/`. See [docs/custom-auth-providers.md](../docs/custom-auth-providers.md).

## Configuration

Config resolution order: YAML file (`CONFIG_FILE` env var → `./config/default.yaml`) with env var overrides (`OSDU_AUTH_CLIENT_ID`, `OSDU_AUTH_CLIENT_SECRET`, `OSDU_AUTH_TENANT_ID`). See `config/azure-example.yaml` for the full schema.

## Relevant Docs

- [docs/custom-auth-providers.md](../docs/custom-auth-providers.md) — Implementing new providers
- [docs/azure-authentication.md](../docs/azure-authentication.md) — Azure-specific auth setup
- [docs/azure-dependencies.md](../docs/azure-dependencies.md) — Why Azure is a separate module
- [docs/authentication-restructure.md](../docs/authentication-restructure.md) — Migration history
- [docs/NOTIFICATION_WEBHOOK.md](../docs/NOTIFICATION_WEBHOOK.md) — Webhook signature verification
