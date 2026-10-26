# Modules

A module is a template file that another template includes. cfnrain inserts the module's resources into the parent template when it packages it. Modules let you split a large template into files and reuse a group of resources in many stacks.

Modules are a client-side feature. CloudFormation only sees the packaged template, so run `cfnrain pkg` to see the result. Run linters such as `cfn-lint` on the packaged template, because they do not know the `Modules` section.

## Example

The module, `bucket-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/docs/bucket-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/docs/module-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/module-expect.yaml"
```

## The Modules section

Each entry in `Modules` has a name and a configuration:

| Key | What it does |
|-----|--------------|
| `Source` | Where the module file is. See [Sources](writing.md#sources). |
| `Properties` | Values for the module's `Parameters` |
| `Overrides` | Changes to the module's resources. See [Overrides](writing.md#overrides). |
| `ForEach` | Make one copy of the module for each item in a list. See [ForEach](foreach.md). |

## Experimental status

The `Modules` section is not experimental and needs no flag. Only the older [`!Rain::Module`](packages.md#the-older-rainmodule-directive) directive needs `--experimental`. See [Experimental features](../concepts/experimental.md).

The feature is still young. The format can change in a minor release.

## Next steps

- [Writing modules](writing.md): parameters, outputs, logical IDs, overrides and sources
- [ForEach](foreach.md): one copy of a module for each item in a list
- [IfParam and conditions](conditions.md): leave out resources based on the properties
- [Intrinsic functions](intrinsics.md): `Fn::Merge`, `Fn::InsertFile` and `Fn::Invoke`
- [Packages](packages.md): share modules as zip files
- [Constants](../constants.md): values that you write once
