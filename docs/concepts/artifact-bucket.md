# The artifact bucket

cfnrain uploads files to an S3 bucket, the artifact bucket. CloudFormation then reads the files from there. The bucket holds:

- the files that [packaging](packaging.md) uploads, for example Lambda code from [`!Rain::S3`](../directives/s3.md) or an [implicit upload](../directives/implicit-uploads.md)
- templates larger than 51,200 bytes, which CloudFormation cannot take inline
- the state files of [`cc`](../commands/stack/cc.md)

## Which bucket cfnrain uses

cfnrain finds the bucket in this order:

1. The `--s3-bucket` flag
2. The SSM parameter `rain-bucket`
3. `rain-artifacts-<account id>-<region>`

cfnrain resolves the bucket on every deploy, also when the template has nothing to upload.

`--s3-prefix` adds a prefix to the object keys. `--s3-owner` sets the account that must own the bucket. cfnrain sends the owner with every upload, so an upload to a bucket that another account owns fails. See [Common flags](../commands/common-flags.md#artifact-bucket).

## When cfnrain creates the bucket

If the bucket does not exist, cfnrain asks to create it. With `-y`, it creates the bucket without a question. To create the bucket ahead of time, run [`cfnrain bootstrap`](../commands/template/bootstrap.md).

cfnrain creates the bucket with these settings:

- AES256 encryption
- All public access blocked
- A lifecycle rule that deletes objects, old object versions and incomplete multipart uploads after 7 days

## Objects expire after 7 days

CloudFormation does not need the uploaded files after a deploy. But if you deploy an old change set or roll back to a template after more than 7 days, the files can be gone. The `cc` state files are in the same bucket and expire too.

If you need to keep objects longer, use `--s3-bucket` with your own bucket.

The ID of the lifecycle rule says `delete after 14 days`, which is wrong. See [Known issues](../known-issues.md#the-artifact-bucket-rule-says-14-days).
