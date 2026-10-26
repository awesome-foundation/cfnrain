# Intrinsic functions

cfnrain adds some intrinsic functions that CloudFormation does not have. cfnrain resolves them when it packages the template, so CloudFormation never sees them. It also simplifies [`Fn::Join` and `Fn::Select`](#fnjoin-and-fnselect) when it can.

| Function | Short form | Where it works |
|----------|------------|----------------|
| [`Fn::Merge`](#fnmerge) | `!Merge` | In module resources |
| [`Fn::InsertFile`](#fninsertfile) | `!InsertFile` | Anywhere |
| [`Fn::Invoke`](#fninvoke) | `!Invoke` | In a template that has a `Modules` section |

!!! note "Fn::Merge only runs in modules"
    In the resources of the parent template, cfnrain leaves `Fn::Merge` as it is. CloudFormation has no `Fn::Merge`, so a deploy with `Fn::Merge` outside a module fails.

## Fn::Merge

`Fn::Merge` merges a list of maps into one map, or a list of lists into one list. For maps, a later key replaces an earlier one. For lists, cfnrain joins the lists and keeps duplicates.

Use it to combine values that come from module properties, such as tags from the parent and tags from the module.

The module:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/merge-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/merge-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/merge-expect.yaml"
```

## Fn::InsertFile

`Fn::InsertFile <path>` inserts the contents of a file as a string. The path is relative to the template. Unlike [`!Rain::Embed`](../directives/embed.md), it keeps the whitespace at the end of the file.

The template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/insertfile-template.yaml"
```

`insertfile-file.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/insertfile-file.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/insertfile-expect.yaml"
```

## Fn::Invoke

`Fn::Invoke` calls a module like a function. It takes three items:

1. The name of a module in the `Modules` section
2. A map of properties. These replace the properties that the `Modules` section sets.
3. The name of a module output

cfnrain evaluates the module with the new properties and inserts the output. A module for `Invoke` does not need resources. `Parameters` and `Outputs` are enough. A common use is a naming module that joins an environment, an application and a resource name.

The naming module, `invoke-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/invoke-module.yaml"
```

A module that returns an object, `invoke-module-b.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/invoke-module-b.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/invoke-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/invoke-expect.yaml"
```

The `Label` module also has a resource, `Why`. cfnrain adds it to the template as `LabelWhy`, with the properties from the `Modules` section. `Fn::Invoke` does not add resources.

## Fn::Join and Fn::Select

cfnrain also simplifies two CloudFormation functions when it can resolve them while it packages the template:

- `Fn::Join` becomes a plain string when every item in the list is a string. This works in the parent template and in modules.
- `Fn::Select` becomes the selected item when the list and the index are known.

!!! bug "Fn::Select takes its arguments in the reverse order"
    cfnrain simplifies `Fn::Select` only when the list comes first and the index second: `Fn::Select: [!Ref Things, 0]`. CloudFormation uses the opposite order: `Fn::Select: [0, !Ref Things]`. If you use the CloudFormation order, cfnrain leaves the function as it is and CloudFormation resolves it. That works. If you use the cfnrain order and cfnrain cannot resolve the list, CloudFormation gets the arguments in the wrong order and the deploy fails. Use the CloudFormation order. See [Known issues](../known-issues.md).

The test for `Fn::Select` uses the cfnrain order:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/select-module.yaml"
```

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/select-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/select-expect.yaml"
```
