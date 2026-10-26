# forecast (experimental)

`cfnrain forecast` looks for things that can make a deploy fail, such as a bucket name that already exists or an EC2 key pair that does not exist. It is not a linter. Use `cfn-lint` to check the syntax. It needs `-x` (`--experimental`).

```sh
cfnrain forecast -x app.yaml app-prod
```

The output in this example is illustrative. When a check fails, it looks like this:

```text
Stormy weather ahead! 🌪

1 checks failed out of 9 total checks
FG001 FAIL on line 6: AWS::S3::Bucket Bucket - Resource with this name already exists
```

When all checks pass, the output starts with `Clear skies!` and the number of checks. The exit status is 1 when a check fails, so you can use `forecast` in CI.

| Flag | What it does |
|------|--------------|
| `--action` | `create`, `update`, `delete` or `all` (default) |
| `-a`, `--all` | Show all checks, not only the failed ones |
| `--include-iam` | Also check IAM permissions. This is slow. |
| `--role-arn` | The role to use for the IAM checks |
| `--type` | Check only one resource type |
| `--ignore` | Skip resource types or check codes, for example `AWS::S3::Bucket,F0002` |
| `--params`, `--tags`, `-c` (`--config`) | Parameter and tag values, as for `deploy` |
| `--plugin` | The path to a forecast plugin `.so` file with more checks. The release binaries cannot load plugins. See [#13](https://github.com/awesome-foundation/cfnrain/issues/13). |
| `--plugin-only` | Run only the plugin checks, not the built-in ones |

The checks cover S3 buckets and bucket policies, EC2 instances, launch templates and security groups, RDS clusters, load balancer listeners and target groups, SNS topics, Lambda functions and SageMaker notebook instances.
