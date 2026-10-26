# bootstrap

`cfnrain bootstrap` creates the artifact bucket. `deploy` and `pkg` create it too when they need it, so you only need `bootstrap` to create it ahead of time.

```sh
cfnrain bootstrap
cfnrain bootstrap -y --s3-bucket my-artifacts
```

`-y` (`--yes`) creates the bucket without asking. `--s3-bucket`, `--s3-prefix` and `--s3-owner` work as for `pkg`. See [Artifact bucket](../../getting-started.md#the-artifact-bucket) for the bucket name and settings.
