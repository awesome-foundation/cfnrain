# AGENTS.md

This file tells coding agents and human contributors how to work in this repository. Read it before you change code.

## What cfnrain is

cfnrain is a command line tool for AWS CloudFormation. It is a community fork of AWS CloudFormation Rain, which AWS archived on 2026-07-29. cfnrain started from Rain's last commit and works as a replacement for Rain v1.24.4.

cfnrain is not affiliated with AWS.

## Goals

1. Keep existing Rain users working. A template, a config file or a script that works with Rain v1.24.4 must work with cfnrain.
2. Keep the tool maintained: dependency updates, security fixes, and bug fixes.
3. Make the documentation good enough that people do not need to read the code.

The project direction is open for discussion in [Direction](https://github.com/awesome-foundation/cfnrain/discussions/20). Do not start a large change, such as removing a command, before it has agreement there or in an issue.

## Repository layout

| Path | Contents |
|---|---|
| `cmd/cfnrain/` | The `main` package for the released binary |
| `internal/cmd/` | One package for each command. `internal/cmd/rain/` is the root command |
| `internal/aws/` | Thin wrappers around the AWS SDK, one package for each service |
| `internal/config/` | The name and the version. Release tooling updates `version.go` |
| `cft/` | The template model: parse, format, diff, graph, and the `pkg` engine |
| `cft/pkg/` | Packaging: the `!Rain::*` directives, modules, constants, and the golden tests in `cft/pkg/tmpl/` |
| `docs/` | The documentation site, as Markdown. `zensical.toml` sets the navigation |
| `test/templates/` | Sample templates that tests use |
| `scripts/` | Helper scripts. `scripts/integ.sh` runs against a real AWS account |

`cmd/cfn-format`, `cmd/cfn-skeleton`, `cmd/aws-console` and `cmd/sample_plugin` came from Rain. The release does not build them.

## Compatibility requirements

These names and behaviors must not change unless a release says so as a breaking change. Users depend on each one.

- The template directives `!Rain::S3`, `!Rain::S3Http`, `!Rain::Embed`, `!Rain::Include`, `!Rain::Env`, `!Rain::Module` and `!Rain::Constant`, and the `${Rain::...}` and `${Const::...}` forms.
- The template sections `Rain`, `Modules`, `Packages` and `Constants`, and `Fn::Merge`, `Fn::InsertFile` and `Fn::Invoke`.
- The environment variables `RAIN_VAR_*`, `RAIN_DEFAULT_TAG_*` and `RAIN_CACHE_DIR`.
- The artifact bucket name `rain-artifacts-<account>-<region>`, the SSM parameter `rain-bucket`, and the SHA-256 object keys.
- `Metadata.Rain` on `AWS::S3::Bucket` resources. The `rm` command reads it.
- The `--config` file format, and every command, flag and alias.

If your change touches one of these, say so in the pull request.

## Build and test

You need the Go version that `go.mod` names.

```sh
go build ./cmd/cfnrain
staticcheck ./...
go vet ./...
go test ./...
```

CI runs the same checks on every pull request (`.github/workflows/test.yml`). Install staticcheck with `go install honnef.co/go/tools/cmd/staticcheck@latest`.

`go test ./...` includes:

- The golden tests in `cft/pkg/tmpl/`. Each `*-template.yaml` file has a `*-expect.yaml` file with the expected `pkg` output. If you change packaging output on purpose, update the expected file in the same commit.
- `TestEveryCommandIsDocumented`. It fails when a command has no mention in `docs/`.

Do not add a unit test that needs AWS credentials.

## Documentation

- The site is the Markdown in `docs/`. Build it with `zensical build --clean --strict`.
- Each fact has one page. Link to that page. Do not copy the fact to a second page.
- When you add or change a command or a flag, update its page in `docs/commands/` in the same pull request.
- When you find a bug while you write docs, open an issue and add an entry that links to it in `docs/known-issues.md`. Do not change behavior in a docs pull request. The pull request that fixes the bug removes the entry.

## Pull requests and commits

- Use [Conventional Commits](https://www.conventionalcommits.org/): `fix(deploy): handle an empty change set`. release-please makes the version number and the changelog from these prefixes. A breaking change needs `!` and a `BREAKING CHANGE:` footer.
- The repository uses squash merges only. The pull request title becomes the commit message, so it follows the same format.
- `main` accepts a pull request only when CI passes.
- Keep one change in each pull request. Do not reformat code that your change does not touch.
- Do not edit `CHANGELOG.md` or the version in `internal/config/version.go`. Release tooling owns them.

## Code style

- Follow the code around your change. Match its names, error handling and comments.
- Run `gofmt` on every Go file you change.
- Some old code panics on errors. New code returns errors.
- Do not add a dependency for a few lines of code.

## Security

Report vulnerabilities as `SECURITY.md` describes. Do not put credentials, account IDs or bucket names from a real account in code, tests or docs.
