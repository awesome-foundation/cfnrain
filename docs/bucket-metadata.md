# S3 bucket metadata (experimental)

An `AWS::S3::Bucket` resource can have a `Metadata.Rain` section. It tells `cfnrain deploy` and `cfnrain rm` to do more work around the stack operation: run build scripts, upload files to the bucket, invalidate a CloudFront distribution, and empty the bucket before a delete.

This feature is experimental. If a bucket in the template has `Metadata.Rain`, `deploy` and `rm` stop with `metadata commands require the --experimental flag` unless you add `--experimental`.

## Example

```yaml
Resources:
  SiteBucket:
    Type: AWS::S3::Bucket
    Metadata:
      Rain:
        EmptyOnDelete: true
        Content: site/dist
        DistributionLogicalId: SiteDistribution
        RunBefore:
          Command: build-site.sh
          Args:
            - production
        RunAfter:
          Command: write-config.sh
          Args:
            - Rain::OutputValue ApiUrl

  SiteDistribution:
    Type: AWS::CloudFront::Distribution
    Properties:
      # ...

Outputs:
  ApiUrl:
    Value: !Sub https://${Api}.execute-api.${AWS::Region}.amazonaws.com/prod
```

```sh
cfnrain deploy site.yaml site-prod --experimental
cfnrain rm site-prod --experimental
```

## Keys

| Key | When | What it does |
|-----|------|--------------|
| `RunBefore` | Before the change set | Runs a command. If it fails, the deploy stops. |
| `RunAfter` | After a successful deploy | Runs a command. It can use stack outputs as arguments. |
| `Content` | After a successful deploy | Uploads every file in a local directory to the bucket |
| `DistributionLogicalId` | After the upload | Invalidates all paths in this CloudFront distribution |
| `EmptyOnDelete` | On `cfnrain rm` | Deletes every object and every object version, so CloudFormation can delete the bucket |
| `Version` | Never | cfnrain ignores it. Change the number to make the template different, so that a deploy runs the steps again. |

### RunBefore and RunAfter

Both take a map with `Command` and an optional `Args` list. `Command` is a path relative to the template file. cfnrain runs it in the template's directory. The command must be executable.

In `RunAfter`, an argument of the form `Rain::OutputValue <OutputKey>` becomes the value of that stack output. Use it to give a build script the URL of an API that the stack created. `RunBefore` cannot use `Rain::OutputValue`, because the stack outputs do not exist yet.

If the command fails, cfnrain prints its output and stops. When it succeeds, cfnrain prints `Successfully ran <command>`.

### Content

`Content` is a directory, relative to the template file. cfnrain uploads each file with a key that is its path inside that directory. `site/dist/css/main.css` becomes `css/main.css`. cfnrain does not delete objects that are no longer in the directory.

An empty value or `RAIN_NO_CONTENT` skips the upload.

### EmptyOnDelete

When `EmptyOnDelete` is `true`, `cfnrain rm` asks for confirmation and then deletes all objects in the bucket, also the old versions. With `-y` it does not ask. Use it for development stacks. Be careful in production.

## Order of steps

1. Package the template
2. `RunBefore`
3. Create the change set, confirm, execute and wait
4. `RunAfter`
5. Upload `Content`
6. Invalidate `DistributionLogicalId`

!!! warning "--detach does not wait for the stack"
    With `--detach`, cfnrain does not wait for the stack. It runs `RunAfter`, the upload and the invalidation right away, while CloudFormation is still working. On a new stack the bucket may not exist yet. Do not use `--detach` with `Metadata.Rain`.

!!! warning "Templates run code"
    `RunBefore` and `RunAfter` run programs from your computer. See [Templates can run code](concepts/packaging.md#templates-can-run-code).

A complete example with modules is in [`test/webapp`](https://github.com/awesome-foundation/cfnrain/tree/main/test/webapp).
