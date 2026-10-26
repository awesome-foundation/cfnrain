# Your first deploy

This page deploys a small stack, changes it, and removes it.

You need cfnrain ([Install](install.md)) and AWS credentials that can create CloudFormation stacks and S3 buckets. Run `cfnrain info` to see the account and region that cfnrain uses.

## Deploy

Save this template as `bucket.yaml`:

```yaml
Parameters:
  Env:
    Type: String
    Description: The environment name
    AllowedValues: [dev, prod]

Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      Tags:
        - Key: Env
          Value: !Ref Env

Outputs:
  BucketName:
    Value: !Ref Bucket
```

Deploy it:

```sh
cfnrain deploy bucket.yaml
```

The first time, cfnrain asks to create its [artifact bucket](concepts/artifact-bucket.md). Answer yes.

cfnrain asks for the value of the `Env` parameter:

```text
Enter a value for parameter 'Env' "The environment name":
```

Enter `dev`. cfnrain then shows what CloudFormation will change, and asks you to confirm:

```text
CloudFormation will make the following changes:
Stack bucket:
  + AWS::S3::Bucket Bucket
Do you wish to continue? (Y/n)
```

Answer yes. cfnrain shows the progress, then prints the outputs. The stack name is `bucket`, the template file name without its extension.

## Change the stack

Add a second tag to the bucket in `bucket.yaml`, and run the same command again:

```sh
cfnrain deploy bucket.yaml
```

cfnrain asks for `Env` again and shows the existing value. Press Enter to keep it. The change set shows `>` for the modified bucket.

## Deploy without questions

Give the parameter on the command line and add `-y`. cfnrain then asks nothing:

```sh
cfnrain deploy bucket.yaml -y --params Env=dev
```

This is how you deploy from scripts and CI. See [CI usage](ci.md).

## Remove the stack

```sh
cfnrain rm bucket
```

cfnrain shows the stack and asks you to confirm. Add `-y` to skip the question.

## Next steps

- [Deploying](concepts/deploying.md) explains the deploy flow, parameter sources and change sets.
- [Template directives](directives/index.md) and [modules](modules/index.md) add features to your templates.
- The [`deploy`](commands/stack/deploy.md) page lists every flag.
