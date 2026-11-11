# Release procedure

Follow the [Release notes rules in AGENTS.md](../AGENTS.md#release-notes).
See https://github.com/awesome-foundation/cfnrain/pull/304 for the beta procedure.

## Before a release

Run these commands from the repository root. The schema refresh needs AWS credentials.

1. Build cfnrain: `go build ./cmd/cfnrain`.
2. Refresh the embedded schemas: `./scripts/cache-schemas.sh internal/aws/cfn/schemas`.
3. Merge the regenerated schemas and resource type list through a reviewed PR.

Remove this schema refresh step when `build` is removed in 2.0: https://github.com/awesome-foundation/cfnrain/issues/299.
