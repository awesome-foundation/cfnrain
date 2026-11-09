# Changelog

## [1.25.5](https://github.com/awesome-foundation/cfnrain/compare/v1.25.4...v1.25.5) (2026-11-09)

This release fixes bugs in `deploy`, `rm`, `pkg`, `fmt`, `merge` and `build`. Most of them came from Rain v1.24.4. Some commands now stop with an error where they used to fail silently or produce bad output. Read **Behavior changes** before you upgrade CI pipelines.

### Behavior changes

- **`deploy` asks before it deletes a failed stack.** When a stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, `deploy` now asks for confirmation before deleting it. The answer defaults to No. Use `-y` to keep the old automatic delete. Without a terminal and without `-y`, `deploy` stops. ([#106](https://github.com/awesome-foundation/cfnrain/pull/106), fixes #43)
- **`deploy` reports a failed final stack lookup.** If the stack deploys but the final lookup fails, `deploy` still enables termination protection and runs the post-deploy steps. It then exits with an error instead of printing an empty summary. ([#116](https://github.com/awesome-foundation/cfnrain/pull/116), fixes #110; follow-up [#130](https://github.com/awesome-foundation/cfnrain/pull/130))
- **`deploy` lists all unknown parameters at once:** `unknown parameter: Alpha, Mike, Zulu`. `--ignore-unknown-params` no longer prints warnings; use `--debug` to see the names. ([#74](https://github.com/awesome-foundation/cfnrain/pull/74), fixes #68)
- **`pkg` fails when an implicit upload fails.** It used to leave the local path in the template and exit 0. It now names the property and the missing file, and exits 1. ([#114](https://github.com/awesome-foundation/cfnrain/pull/114), fixes #44)
- **`pkg` rejects a Rain constant that it cannot insert correctly.** These used to package to an empty string. `pkg` now stops with an error that names the constant ([#125](https://github.com/awesome-foundation/cfnrain/pull/125); follow-ups [#132](https://github.com/awesome-foundation/cfnrain/pull/132), [#133](https://github.com/awesome-foundation/cfnrain/pull/133)):
  - a constant that is a long-form `!Sub`, a `!Ref`, an object or a list, used inside a `!Sub`;
  - a plain constant that refers to a `!Sub` constant, such as `C: "${Rain::B}"` with `B: !Sub ${Env}-b`, when the template uses it. The error tells you to use `!Sub` for the outer value, or `!Rain::Constant B` for the whole value.
- **`pkg` rejects an out-of-range literal `Fn::Select` index** in CloudFormation argument order, such as `!Select [5, [a, b]]`. CloudFormation rejects this index too. ([#123](https://github.com/awesome-foundation/cfnrain/pull/123))
- **`build` rejects `--no-cache` together with `--only-cache`.** ([#136](https://github.com/awesome-foundation/cfnrain/pull/136))
- **Help and flag errors changed:** pflag 1.1 changed how help and flag errors look. Boolean flags now show `[=true|false]` in help, and error text for invalid flag values is different. ([#109](https://github.com/awesome-foundation/cfnrain/pull/109))

### deploy and rm

- `rm` no longer reports a failure when the stack is deleted before its first status check. ([#105](https://github.com/awesome-foundation/cfnrain/pull/105), fixes #95)
- Stack lookups no longer panic when CloudFormation returns an empty stack list. ([#105](https://github.com/awesome-foundation/cfnrain/pull/105), fixes #93)
- After `deploy` deletes a failed stack, it waits on the stack ID, so a fast delete no longer makes it fail. ([#116](https://github.com/awesome-foundation/cfnrain/pull/116))
- `rm` no longer panics on a stack in `REVIEW_IN_PROGRESS`, such as one that `deploy --no-exec` created, or when the response has no termination protection value. ([#73](https://github.com/awesome-foundation/cfnrain/pull/73), fixes #67; [#98](https://github.com/awesome-foundation/cfnrain/pull/98), fixes #94)

### pkg

- A Rain constant that is itself a `!Sub` now works inside another `!Sub`, instead of packaging to an empty string. Escaped `${!Name}` values keep their `Fn::Sub`, so CloudFormation removes the `!` as it should. ([#125](https://github.com/awesome-foundation/cfnrain/pull/125), fixes #30)
- `Fn::Select` now resolves in CloudFormation's argument order, `[index, list]`, in templates and in modules. The old reverse order still works. Lists computed by CloudFormation, such as `!GetAZs ''`, are left untouched. ([#123](https://github.com/awesome-foundation/cfnrain/pull/123), fixes #47; follow-up [#129](https://github.com/awesome-foundation/cfnrain/pull/129))
- `!Rain::S3` `Run` fixes ([#61](https://github.com/awesome-foundation/cfnrain/pull/61), [#88](https://github.com/awesome-foundation/cfnrain/pull/88), fixes #60):
  - `Run` finds the script when you give the template as an absolute path.
  - An absolute `Run` path is used as given.
  - Two directives with the same `Path` and different `Run` scripts now each upload their own build, instead of both getting the first one.
  - A failed script shows why it failed.

### fmt

- `fmt` keeps comments on function keys when it writes the short form, such as `!Sub`. ([#112](https://github.com/awesome-foundation/cfnrain/pull/112), fixes #85)
- `fmt` keeps an anchor on a function argument instead of failing. ([#113](https://github.com/awesome-foundation/cfnrain/pull/113), fixes #83)
- `fmt` keeps YAML merge keys (`<<: *base`), including in flow style. `--json` rejects them with a clear error, because JSON has no merge keys. ([#115](https://github.com/awesome-foundation/cfnrain/pull/115), fixes #84)
- `fmt` keeps comments in templates with CRLF line endings. ([#75](https://github.com/awesome-foundation/cfnrain/pull/75), fixes #64)
- cfnrain now understands ten more short-form tags: `!Contains`, `!EachMemberEquals`, `!EachMemberIn`, `!RefAll`, `!ValueOf`, `!ValueOfAll`, `!Length`, `!ToJsonString`, `!GetStackOutput` and `!Transform`. JSON output and `diff` see these functions, and `fmt` keeps the form the template uses. ([#80](https://github.com/awesome-foundation/cfnrain/pull/80), fixes #63, #81)

### merge

- `merge` no longer panics on a file it cannot parse. It names the file and shows the parse error. ([#111](https://github.com/awesome-foundation/cfnrain/pull/111), fixes #66)

### build

- `build --only-cache` now uses only the schemas built into cfnrain and makes no AWS calls. It used to do the same as `--no-cache`. For a resource type that is not built in, it names the type in the error. ([#136](https://github.com/awesome-foundation/cfnrain/pull/136), fixes #41)
- The templates that `build` generates, and the shipped `compliant-bucket` module, use the correct IAM action `s3:ReplicateTags` in replication policies. The old name, `s3:ReplicationTags`, is not an IAM action. ([#100](https://github.com/awesome-foundation/cfnrain/pull/100))
- The `webapp` template that `build` generates no longer sets `AuthorizerId` to the literal string `AWS::NoValue` on its JWT method. ([#96](https://github.com/awesome-foundation/cfnrain/pull/96))

### Help text and generated resources

- The `pkg` help now lists the correct S3 `Format` values (`URI`, `Http` and `Object`), the real defaults, and the `Run` and `Extension` options. ([#119](https://github.com/awesome-foundation/cfnrain/pull/119), fixes #45; follow-up [#131](https://github.com/awesome-foundation/cfnrain/pull/131))
- The `stackset deploy` help no longer says the template file is optional. ([#118](https://github.com/awesome-foundation/cfnrain/pull/118), fixes #46)
- New artifact buckets get the lifecycle rule ID "delete after 7 days", which matches the expiry. Existing buckets don't change. ([#117](https://github.com/awesome-foundation/cfnrain/pull/117), fixes #42)

### Documentation

- New page: [Why CloudFormation and Rain](https://awesome-foundation.github.io/cfnrain/why.html). ([#58](https://github.com/awesome-foundation/cfnrain/pull/58), [#59](https://github.com/awesome-foundation/cfnrain/pull/59))
- The known issues page now lists only issues that the project will not fix. Open bugs are tracked in GitHub issues, and their workarounds are on the command pages. ([#134](https://github.com/awesome-foundation/cfnrain/pull/134), [#135](https://github.com/awesome-foundation/cfnrain/pull/135))

### Maintenance

- Updated the Go dependencies, including the AWS SDK and pflag 1.1.0. ([#62](https://github.com/awesome-foundation/cfnrain/pull/62), [#108](https://github.com/awesome-foundation/cfnrain/pull/108), [#109](https://github.com/awesome-foundation/cfnrain/pull/109))
- The release job builds the binaries without a restored Go cache. ([#101](https://github.com/awesome-foundation/cfnrain/pull/101))
- CI now runs the integration tests against AWS, including stack lifecycle and artifact checks. ([#96](https://github.com/awesome-foundation/cfnrain/pull/96), [#102](https://github.com/awesome-foundation/cfnrain/pull/102))
- Fixed a flaky table test. ([#124](https://github.com/awesome-foundation/cfnrain/pull/124), fixes #86)

## [1.25.4](https://github.com/awesome-foundation/cfnrain/compare/v1.25.3...v1.25.4) (2026-10-27)


### Bug Fixes

* **pkg:** reject zip entries outside the destination ([89e08ef](https://github.com/awesome-foundation/cfnrain/commit/89e08ef34f382aa4ee3eb0ddaa5adc9b5e96065c))

## [1.25.3](https://github.com/awesome-foundation/cfnrain/compare/v1.25.2...v1.25.3) (2026-10-27)


### Bug Fixes

* **deps:** bump pkl-go, go-runewidth and testify ([#27](https://github.com/awesome-foundation/cfnrain/issues/27)) ([d560313](https://github.com/awesome-foundation/cfnrain/commit/d5603133f2fd9fc0b54507a6aa536924332e41f2))

## [1.25.2](https://github.com/awesome-foundation/cfnrain/compare/v1.25.1...v1.25.2) (2026-10-26)


### Bug Fixes

* **deps:** update Go dependencies ([#18](https://github.com/awesome-foundation/cfnrain/issues/18)) ([3c50f7c](https://github.com/awesome-foundation/cfnrain/commit/3c50f7c9a0d992323846baf0225bebf5680aae3c))

## [1.25.1](https://github.com/awesome-foundation/cfnrain/compare/v1.25.0...v1.25.1) (2026-10-26)


### Continuous Integration

* publish releases as drafts first and build fewer targets at once ([#21](https://github.com/awesome-foundation/cfnrain/issues/21)) ([c97fb21](https://github.com/awesome-foundation/cfnrain/commit/c97fb21e7d4a07852c1b37fc1062df712a820db1))

## [1.25.0](https://github.com/awesome-foundation/cfnrain/compare/v1.24.4...v1.25.0) (2026-10-26)


### Features

* rename the project to cfnrain ([#2](https://github.com/awesome-foundation/cfnrain/issues/2)) ([2f41823](https://github.com/awesome-foundation/cfnrain/commit/2f41823dfa55aff8de965e4c725dba8097a32059))
* stop writing AWS usage metrics into packaged templates ([#3](https://github.com/awesome-foundation/cfnrain/issues/3)) ([c0f97f3](https://github.com/awesome-foundation/cfnrain/commit/c0f97f387f06ce439034c97c762b8d25f8af89fa))
