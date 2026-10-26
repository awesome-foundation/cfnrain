# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead.

## A !Sub constant inside another !Sub packages to an empty value

A `Rain:` constant that is itself a `!Sub` string, used with `${Rain::<Name>}` inside another `!Sub`, packages to an empty string. There is no error.

Use `!Rain::Constant <Name>` for that constant instead. See [Constants](directives/constants.md#known-issue-a-sub-constant-inside-another-sub).

## Fn::Select takes its arguments in the reverse order

cfnrain simplifies `Fn::Select` only when the list comes first and the index second. CloudFormation uses the index first. Write `Fn::Select` in the CloudFormation order. cfnrain then leaves it for CloudFormation, which resolves it. See [Intrinsic functions](modules/intrinsics.md#fnjoin-and-fnselect).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`.

## The artifact bucket deletes objects after 7 days, not 14

When cfnrain creates the artifact bucket, it adds a lifecycle rule with the ID `delete after 14 days`. The rule deletes objects, old object versions and incomplete multipart uploads after 7 days.

CloudFormation does not need the uploaded files after a deploy. But if you deploy an old change set or roll back to a template after more than 7 days, the files can be gone. The `cc` state files are in the same bucket and expire too. Use `--s3-bucket` with your own bucket if you need to keep objects longer.

## info --creds prints the secret key

`cfnrain info --creds` prints the `SecretAccessKey` and the `SessionToken` in plain text. Do not run it where other people or logs can see the output.

## Failed stacks are deleted without a prompt

If a stack is in `ROLLBACK_COMPLETE`, `CREATE_FAILED` or `REVIEW_IN_PROGRESS`, `cfnrain deploy` deletes it before it deploys, also without `-y`. See [Getting started](getting-started.md#change-the-stack).

## Implicit uploads fail without an error

If an [implicit upload](directives/implicit-uploads.md) fails, for example because the local path does not exist, cfnrain leaves the path in the template and continues. CloudFormation then fails on the path. Run `cfnrain pkg --debug` to see the reason.

## Things that do not exist

Older Rain documentation describes these, but the code does not have them:

- `SourceMap` metadata on packaged resources
- A `VersionProperty` option for `!Rain::S3`
- Pkl input for `diff`, `merge` and `tree`
- The `$MapValue` placeholder in `ForEach`
- The `ModuleExtension` resource and `Metadata.Extends` for `!Rain::Module`

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
