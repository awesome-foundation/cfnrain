# Constants and !Rain::Constant

The `Rain:` section of a template can hold constants. A constant is a value that you write once and use in many places. cfnrain removes the `Rain:` section when it packages the template.

```yaml
Rain:
  Constants:
    App: billing
```

Use a constant in two ways:

- `!Rain::Constant <Name>` inserts the value. The value can be a string or an object.
- `${Rain::<Name>}` inside a `!Sub` string inserts a string value.

A constant can use the constants above it, with `${Rain::<Name>}` in a `!Sub`. If a `!Sub` has no variables left after cfnrain inserts the constants, cfnrain replaces it with a plain string.

## Example

The template:

```yaml
--8<-- "cft/pkg/tmpl/docs/constants-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/constants-expect.yaml"
```

`${Env}` stays in the output, because it is a template parameter that CloudFormation resolves.

## Known issue: a !Sub constant inside another !Sub

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

## The Constants section

Templates and modules can also have a top-level `Constants:` section, with `${Const::<Name>}` and `!Ref Const::<Name>`. That section belongs to the module system. See [Constants section](../modules/constants.md).
