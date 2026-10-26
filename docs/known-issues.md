# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead.

## A !Sub constant inside another !Sub packages to an empty value

A `Rain:` constant that is itself a `!Sub` string, used with `${Rain::<Name>}` inside another `!Sub`, packages to an empty string. There is no error.

Use `!Rain::Constant <Name>` for that constant instead. See [Constants](constants.md#known-issue-a-sub-constant-inside-another-sub).

## Fn::Select takes its arguments in the reverse order

cfnrain simplifies `Fn::Select` only when the list comes first and the index second. CloudFormation uses the index first. Write `Fn::Select` in the CloudFormation order. cfnrain then leaves it for CloudFormation, which resolves it. See [Intrinsic functions](modules/intrinsics.md#fnjoin-and-fnselect).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`.

## The artifact bucket rule says 14 days

When cfnrain creates the artifact bucket, it adds a lifecycle rule with the ID `delete after 14 days`. The rule deletes objects after 7 days. See [Objects expire after 7 days](concepts/artifact-bucket.md#objects-expire-after-7-days).

## Failed stacks are deleted without a prompt

If a stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, `cfnrain deploy` deletes it before it deploys, also without `-y`. See [Updates](concepts/deploying.md#updates).

## Implicit uploads fail without an error

If an [implicit upload](directives/implicit-uploads.md) fails, for example because the local path does not exist, cfnrain leaves the path in the template and continues. CloudFormation then fails on the path. Run `cfnrain pkg --debug` to see the reason.

## The pkg help text lists the wrong S3 Format values

`cfnrain pkg --help` lists `Format: Uri|Http`. The code expects `URI`, `Http` or `Object`, and `Object` is the default when you set both `BucketProperty` and `KeyProperty`. With `Format: Uri`, packaging stops with an error. The help text also leaves out the `Run` and `Extension` options. See [S3 and S3Http](directives/s3.md#format) for the correct values.

## The stackset deploy help text says the template is optional

`cfnrain stackset deploy --help` says that without a template file, cfnrain adds a new instance to an existing stack set. The code always reads the first argument as the template file, so you must give one.

## Mistakes in the Rain documentation

The Rain documentation at `aws-cloudformation.github.io/rain` is frozen, and some of it is wrong:

- On `deploy`, `-x` is `--no-exec`, not `--experimental`. See [Experimental features](concepts/experimental.md#-x-means-no-exec-on-deploy).
- The `Modules` section does not need `--experimental`. Only `!Rain::Module` and bucket metadata do.
- `!Rain::S3` takes `Format: URI`, not `Uri`. See [above](#the-pkg-help-text-lists-the-wrong-s3-format-values).
- The `ForEach` placeholders are `$Identifier` and `$Index`. There is no `$MapValue`. See [ForEach](modules/foreach.md).
- `!Rain::Module` does not need a `ModuleExtension` resource or `Metadata.Extends`.
- Packaged resources get no `SourceMap` metadata.
- `!Rain::S3` has no `VersionProperty` option.
- `diff`, `merge` and `tree` do not read Pkl files.

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
