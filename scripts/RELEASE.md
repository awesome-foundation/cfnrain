# Release procedure

Follow the [Release notes rules in AGENTS.md](../AGENTS.md#release-notes).
See https://github.com/awesome-foundation/cfnrain/pull/304 for the beta procedure.

## Before a release

Run these commands from the repository root. The schema refresh needs AWS credentials.

1. Build cfnrain: `go build ./cmd/cfnrain`.
2. Refresh the embedded schemas: `./scripts/cache-schemas.sh internal/aws/cfn/schemas`.
3. Merge the regenerated schemas and resource type list through a reviewed PR.
4. Before merging the release PR, check the latest "Integration tests" run (`.github/workflows/integ.yml`) on `main`:
   `gh run list --workflow integ.yml --branch main --limit 1 --json headSha,status,conclusion`.
   Confirm that it passed for the current `main` commit.
   Do not merge while it is pending, failed, or for an older commit.

Remove this schema refresh step when `build` is removed in 2.0: https://github.com/awesome-foundation/cfnrain/issues/299.
