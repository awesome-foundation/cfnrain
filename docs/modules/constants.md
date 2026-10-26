# Constants section

A template or a module can have a top-level `Constants:` section. A constant is a string or an object that you write once and use in many places. cfnrain replaces the constants and removes the section when it packages the template.

```yaml
Constants:
  S3Arn: "arn:${AWS::Partition}:s3:::"
  BucketName: "${Name}-${AWS::Region}-${AWS::AccountId}"
  BucketArn: "${Const::S3Arn}${Const::BucketName}"
  Tags:
    Team: platform
```

Use a constant in two ways:

- `${Const::<Name>}` inside a `!Sub` string inserts a string constant.
- `!Ref Const::<Name>` inserts an object constant. You can only insert the whole object, not a part of it.

A constant can use the constants above it with `${Const::<Name>}`. cfnrain replaces the constants before it processes the modules. If a `!Sub` has no variables left after that, cfnrain replaces it with a plain string.

## Example

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/constant-template.yaml"
```

The module, `constant-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/constant-module.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/constant-expect.yaml"
```

The parent's `${Const::yyy}` becomes the plain string `zzz`. In the module, `${Const::sub}` keeps `${AWS::AccountId}`, so it stays a `Fn::Sub`.

## Two constant systems

cfnrain has two ways to write constants. Both work.

| | `Constants:` section | `Rain:` section |
|-|----------------------|-----------------|
| Where | Top level of a template or module | `Rain: Constants:` |
| In a `!Sub` | `${Const::Name}` | `${Rain::Name}` |
| Insert a value | `!Ref Const::Name` | `!Rain::Constant Name` |
| Page | This page | [Constants and !Rain::Constant](../directives/constants.md) |

cfnrain replaces both `${Rain::Name}` and `${Const::Name}` in `!Sub` strings. Pick one system per template, so readers know where to look.
