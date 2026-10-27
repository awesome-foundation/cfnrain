# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead, and links to its issue. When a fix ships, its entry goes away.

## A !Sub constant inside another !Sub packages to an empty value

A `Rain:` constant that is itself a `!Sub` string, used with `${Rain::<Name>}` inside another `!Sub`, packages to an empty string. There is no error.

Use `!Rain::Constant <Name>` for that constant instead. See [Constants](constants.md#known-issue-a-sub-constant-inside-another-sub). Issue: [#30](https://github.com/awesome-foundation/cfnrain/issues/30).

## Fn::Select takes its arguments in the reverse order

cfnrain simplifies `Fn::Select` only when the list comes first and the index second. CloudFormation uses the index first. Write `Fn::Select` in the CloudFormation order. cfnrain then leaves it for CloudFormation, which resolves it. See [Intrinsic functions](modules/intrinsics.md#fnjoin-and-fnselect). Issue: [#47](https://github.com/awesome-foundation/cfnrain/issues/47).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`. Issue: [#41](https://github.com/awesome-foundation/cfnrain/issues/41).

## The artifact bucket rule says 14 days

When cfnrain creates the artifact bucket, it adds a lifecycle rule with the ID `delete after 14 days`. The rule deletes objects after 7 days. See [Objects expire after 7 days](concepts/artifact-bucket.md#objects-expire-after-7-days). Issue: [#42](https://github.com/awesome-foundation/cfnrain/issues/42).

## Failed stacks are deleted without a prompt

If a stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, `cfnrain deploy` deletes it before it deploys, also without `-y`. See [Updates](concepts/deploying.md#updates). Issue: [#43](https://github.com/awesome-foundation/cfnrain/issues/43).

## Implicit uploads fail without an error

If an [implicit upload](directives/implicit-uploads.md) fails, for example because the local path does not exist, cfnrain leaves the path in the template and continues. CloudFormation then fails on the path. Run `cfnrain pkg --debug` to see the reason. Issue: [#44](https://github.com/awesome-foundation/cfnrain/issues/44).

## The pkg help text lists the wrong S3 Format values

`cfnrain pkg --help` lists `Format: Uri|Http`. The code expects `URI`, `Http` or `Object`, and `Object` is the default when you set both `BucketProperty` and `KeyProperty`. With `Format: Uri`, packaging stops with an error. The help text also leaves out the `Run` and `Extension` options. See [S3 and S3Http](directives/s3.md#format) for the correct values. Issue: [#45](https://github.com/awesome-foundation/cfnrain/issues/45).

## The stackset deploy help text says the template is optional

`cfnrain stackset deploy --help` says that without a template file, cfnrain adds a new instance to an existing stack set. The code always reads the first argument as the template file, so you must give one. Issue: [#46](https://github.com/awesome-foundation/cfnrain/issues/46).

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
