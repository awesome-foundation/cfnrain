# Constants

A constant is a value that you write once and use in many places. cfnrain replaces the constants and removes their section when it [packages](concepts/packaging.md) the template.

cfnrain has two ways to write constants. Both work. Pick one system per template, so readers know where to look.

| | `Constants:` section | `Rain:` section |
|-|----------------------|-----------------|
| Where | Top level of a template or module | `Rain: Constants:` |
| In a `!Sub` | `${Const::Name}` | `${Rain::Name}` |
| Insert a value | `!Ref Const::Name` | `!Rain::Constant Name` |

cfnrain replaces both `${Rain::Name}` and `${Const::Name}` in `!Sub` strings. If a `!Sub` has no variables left after cfnrain inserts the constants, cfnrain replaces it with a plain string.

## The Constants section

A template or a [module](modules/index.md) can have a top-level `Constants:` section. A constant is a string or an object.

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

A constant can use the constants above it with `${Const::<Name>}`. cfnrain replaces the constants before it processes the modules.

### Example

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

## The Rain section

The `Rain:` section of a template can also hold constants:

```yaml
Rain:
  Constants:
    App: billing
```

Use a constant in two ways:

- `!Rain::Constant <Name>` inserts the value. The value can be a string or an object.
- `${Rain::<Name>}` inside a `!Sub` string inserts a string value.

A constant can use the constants above it, with `${Rain::<Name>}` in a `!Sub`.

### Example

The template:

```yaml
--8<-- "cft/pkg/tmpl/docs/constants-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/constants-expect.yaml"
```

`${Env}` stays in the output, because it is a template parameter that CloudFormation resolves.

### Known issue: a !Sub constant inside another !Sub

!!! bug "This packages to an empty value"
    If a constant is itself a `!Sub` string, do not use it with `${Rain::<Name>}` inside another `!Sub`. cfnrain packages it to an empty string and shows no error.

    ```yaml
    Rain:
      Constants:
        Prefix: !Sub ${Env}-billing
    Resources:
      Table:
        Type: AWS::DynamoDB::Table
        Properties:
          TableName: !Sub ${Rain::Prefix}-orders   # packages to "-orders"
    ```

    Use `!Rain::Constant Prefix` instead, as in the example above. That works.
