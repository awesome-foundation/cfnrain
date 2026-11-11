# Changelog

## [1.26.0-beta.2](https://github.com/awesome-foundation/cfnrain/compare/v1.26.0-beta.1...v1.26.0-beta.2) (2026-11-11)


### Bug Fixes

* **fmt:** keep commented explicit flow keys in the original style ([#305](https://github.com/awesome-foundation/cfnrain/issues/305)) ([2da36ef](https://github.com/awesome-foundation/cfnrain/commit/2da36efd29b2dd5a28a291311d554fe8099297d0))


### Miscellaneous Chores

* release 1.26.0-beta.2 ([#310](https://github.com/awesome-foundation/cfnrain/issues/310)) ([b8efc86](https://github.com/awesome-foundation/cfnrain/commit/b8efc8613931e5acf1caa04adf1a059913852426))

## [1.26.0-beta.1](https://github.com/awesome-foundation/cfnrain/compare/v1.25.5...v1.26.0-beta.1) (2026-11-11)


### Features

* **deploy:** add an opt-in --express flag for Express mode ([#77](https://github.com/awesome-foundation/cfnrain/issues/77)) ([652a459](https://github.com/awesome-foundation/cfnrain/commit/652a459e096342440d7114f1e2e0f80eb63a47c8))
* **deploy:** show Early Validation details when a change set fails ([#76](https://github.com/awesome-foundation/cfnrain/issues/76)) ([6f89382](https://github.com/awesome-foundation/cfnrain/commit/6f893828886eec1abbc03a667068cacbadf60083))
* deprecate the forecast, cc and build commands ([#300](https://github.com/awesome-foundation/cfnrain/issues/300)) ([7323f57](https://github.com/awesome-foundation/cfnrain/commit/7323f579d4d79b4632c622adf66427f286e049d9))
* read a template from stdin when the file name is - ([#78](https://github.com/awesome-foundation/cfnrain/issues/78)) ([94d627d](https://github.com/awesome-foundation/cfnrain/commit/94d627dc286bef8e6da3fba6bacf86a239312d26))


### Bug Fixes

* **fmt:** accept the documented strict-boolean style ([#192](https://github.com/awesome-foundation/cfnrain/issues/192)) ([2b06fe7](https://github.com/awesome-foundation/cfnrain/commit/2b06fe7f4897816777b91baef1f928169026850d))
* **fmt:** accept unchanged YAML NaN values ([#173](https://github.com/awesome-foundation/cfnrain/issues/173)) ([a441179](https://github.com/awesome-foundation/cfnrain/commit/a44117915ed4fd1cee57dda89cc1b21f006b57fb))
* **fmt:** keep comments on empty null items ([#165](https://github.com/awesome-foundation/cfnrain/issues/165)) ([989100c](https://github.com/awesome-foundation/cfnrain/commit/989100c6b493962180f5fc0b16370d41490c1b21))
* **fmt:** keep empty function arguments as strings in JSON output ([#140](https://github.com/awesome-foundation/cfnrain/issues/140)) ([947b67f](https://github.com/awesome-foundation/cfnrain/commit/947b67f6940f495a47aafa04409c8faec793560c))
* **fmt:** keep source quotes on values YAML 1.1 reads as booleans ([#79](https://github.com/awesome-foundation/cfnrain/issues/79)) ([3638c9a](https://github.com/awesome-foundation/cfnrain/commit/3638c9a8cdbbf785f1b93433210f36eb8d92bb0a))
* **fmt:** preserve aliases of anchored functions ([#152](https://github.com/awesome-foundation/cfnrain/issues/152)) ([433ee3a](https://github.com/awesome-foundation/cfnrain/commit/433ee3aa664b05570ae643bf75aea9863acbbfda))
* **fmt:** preserve ambiguous string function arguments ([#168](https://github.com/awesome-foundation/cfnrain/issues/168)) ([9675582](https://github.com/awesome-foundation/cfnrain/commit/9675582c47fd95a2fa5d5e56cd48ca3d3be5c7eb))
* **fmt:** preserve blank lines and tabs in multiline strings ([#191](https://github.com/awesome-foundation/cfnrain/issues/191)) ([77b3339](https://github.com/awesome-foundation/cfnrain/commit/77b333905f8f0f1154498ac444e7e3a618dcbaa6))
* **fmt:** preserve comments after flow collections ([#160](https://github.com/awesome-foundation/cfnrain/issues/160)) ([317871f](https://github.com/awesome-foundation/cfnrain/commit/317871fe19862e8e44b73fdcc89a6eb69a73bfd0))
* **fmt:** preserve comments on empty short-form arguments ([#150](https://github.com/awesome-foundation/cfnrain/issues/150)) ([77054c5](https://github.com/awesome-foundation/cfnrain/commit/77054c5c7e41870b5fb4af12b514e24c76156d50))
* **fmt:** preserve comments on empty values ([#187](https://github.com/awesome-foundation/cfnrain/issues/187)) ([8eec06a](https://github.com/awesome-foundation/cfnrain/commit/8eec06a991e649dea5d74d6c4cd90deaa9cda3ad))
* **fmt:** preserve comments on functions inside lists ([#143](https://github.com/awesome-foundation/cfnrain/issues/143)) ([23f0ace](https://github.com/awesome-foundation/cfnrain/commit/23f0ace7c3d2b2c2fda6f86847dcc4d5a0735b99))
* **fmt:** preserve indented strings in folded style ([#199](https://github.com/awesome-foundation/cfnrain/issues/199)) ([1f1b7d8](https://github.com/awesome-foundation/cfnrain/commit/1f1b7d8589673618eb3c1d34458071eb68ffe5b7))
* **fmt:** preserve inferred tags on nested function arguments ([#193](https://github.com/awesome-foundation/cfnrain/issues/193)) ([4d16079](https://github.com/awesome-foundation/cfnrain/commit/4d16079ca43a0f75a11bcfb369f10eb0bfaa2d3d))
* **fmt:** preserve multiline strings in singlequoted style ([#171](https://github.com/awesome-foundation/cfnrain/issues/171)) ([be612fb](https://github.com/awesome-foundation/cfnrain/commit/be612fbe81679879b1a12e869e77172f5d1a7c09))
* **fmt:** preserve scalar types in literal and folded styles ([#177](https://github.com/awesome-foundation/cfnrain/issues/177)) ([f272a6f](https://github.com/awesome-foundation/cfnrain/commit/f272a6f7f2550e3a5c0290f686e057fa53aff661))
* **fmt:** preserve scalar types with quoted node styles ([#162](https://github.com/awesome-foundation/cfnrain/issues/162)) ([4679dd0](https://github.com/awesome-foundation/cfnrain/commit/4679dd0147e073c60f21a922abba3c688514a6ff))
* **fmt:** preserve sorted flow mapping styles ([#197](https://github.com/awesome-foundation/cfnrain/issues/197)) ([2fd47d2](https://github.com/awesome-foundation/cfnrain/commit/2fd47d20cb0d15c538a5f7f34e9d21099f148165))
* **fmt:** preserve sorted mapping styles and comments ([#169](https://github.com/awesome-foundation/cfnrain/issues/169)) ([61be281](https://github.com/awesome-foundation/cfnrain/commit/61be28118a5c1e80018a408a6b5165629aa008db))
* **fmt:** preserve unsorted key comments in flow and tagged styles ([#198](https://github.com/awesome-foundation/cfnrain/issues/198)) ([ed27b63](https://github.com/awesome-foundation/cfnrain/commit/ed27b63278f4bbf87787cfadc3abc9185b94172c))
* **fmt:** quote strings that start with a tab instead of writing a block ([#211](https://github.com/awesome-foundation/cfnrain/issues/211)) ([2760bef](https://github.com/awesome-foundation/cfnrain/commit/2760befb67f5d20e50993c0424a0983ce346db4e))
* **fmt:** reject YAML alias cycles before formatting ([#204](https://github.com/awesome-foundation/cfnrain/issues/204)) ([77035db](https://github.com/awesome-foundation/cfnrain/commit/77035dbdc2880956c8a2a0051a41ea374c81f590))
* **pkg:** finish YAML traversal before mutating matches ([#161](https://github.com/awesome-foundation/cfnrain/issues/161)) ([b0110f0](https://github.com/awesome-foundation/cfnrain/commit/b0110f03d82eade6e321c40a71cfb8238d22c83a))
* **pkg:** include extension in upload cache key ([#145](https://github.com/awesome-foundation/cfnrain/issues/145)) ([e4846a7](https://github.com/awesome-foundation/cfnrain/commit/e4846a759c9f69d6f8d4d13128f1ae1c365e957c))
* **pkg:** preserve alias cycles while cloning nodes ([#208](https://github.com/awesome-foundation/cfnrain/issues/208)) ([b071607](https://github.com/awesome-foundation/cfnrain/commit/b0716077cbf746d947dec4c73850128967fa646f))
* **pkg:** preserve anchors when intrinsic functions resolve ([#163](https://github.com/awesome-foundation/cfnrain/issues/163)) ([fb5cd2b](https://github.com/awesome-foundation/cfnrain/commit/fb5cd2bf3fa2ff263c227eb5c783f79c07149437))
* **pkg:** preserve empty Fn::Sub functions ([#153](https://github.com/awesome-foundation/cfnrain/issues/153)) ([977b029](https://github.com/awesome-foundation/cfnrain/commit/977b0294e6b1eb1ad41ce13d0f7e3e4185b8d483))
* **pkg:** preserve escaped Fn::Sub variables in modules ([#142](https://github.com/awesome-foundation/cfnrain/issues/142)) ([268c9ac](https://github.com/awesome-foundation/cfnrain/commit/268c9ac62babb2cbeaa91815ae34ec041bd18c32))
* **pkg:** preserve parent anchor bindings across module expansion ([#181](https://github.com/awesome-foundation/cfnrain/issues/181)) ([2b631ca](https://github.com/awesome-foundation/cfnrain/commit/2b631ca092a66dbedbd06daf4b94b56a12fc3e6f))
* **pkg:** preserve parent scope in module parameter expressions ([#176](https://github.com/awesome-foundation/cfnrain/issues/176)) ([135d41c](https://github.com/awesome-foundation/cfnrain/commit/135d41c84f6219b1571e4be18b66f4df6f6ca32a))
* **pkg:** preserve sequence-form Sub module outputs ([#151](https://github.com/awesome-foundation/cfnrain/issues/151)) ([0fb3f9a](https://github.com/awesome-foundation/cfnrain/commit/0fb3f9a2164a5fcf5248e495a5b977c841604192))
* **pkg:** replace constants in sequence-form Sub ([#141](https://github.com/awesome-foundation/cfnrain/issues/141)) ([90895ce](https://github.com/awesome-foundation/cfnrain/commit/90895ce19f7565fe4b34ef2a670735c6badb1c9e))
* **stackset:** correct the instance key in the sample config ([#238](https://github.com/awesome-foundation/cfnrain/issues/238)) ([2840e50](https://github.com/awesome-foundation/cfnrain/commit/2840e508343a5ecc8ed12ecd5eaafe62adc367a4))


### Miscellaneous Chores

* release 1.26.0-beta.1 ([#306](https://github.com/awesome-foundation/cfnrain/issues/306)) ([161df88](https://github.com/awesome-foundation/cfnrain/commit/161df880a5faeb5bae2069471940163192f0b163))

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
