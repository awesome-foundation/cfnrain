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

## File paths in module resources

A relative path uses the directory of the file that supplies its value. This applies to `!Rain::S3` (`Path` and `Run`), `!Rain::S3Http`, `!Rain::Embed`, `!Rain::Include`, and `Fn::InsertFile`. Module resources and outputs use the same rule. Absolute paths do not change.

- A path written in a local module, including a parameter default, uses the module file's directory.
- A path passed through `Properties` or `Overrides` keeps the directory of the file that supplied it. Passing the value through nested modules does not change that directory.
- A `Run` script runs in the directory of the file that supplies `Run`. `Path` can come from a different file.
- A local `Source`, including a nested `Source` or `!Rain::Module` source, uses the directory of the file that supplies it.
- Modules loaded from zip packages, HTTPS, or S3 have no local module directory. Their local asset paths keep the parent template's directory. The shell's working directory does not change this rule.

For example, `modules/worker.yaml` with `Run: build.sh` and `Path: dist` uses `modules/build.sh` and `modules/dist`. If the parent passes `BuildScript: buildapi.sh` through `Properties`, `Run: !Ref BuildScript` uses the parent's `buildapi.sh`.

Module constants resolve in the module that defines them. A file directive stored in a constant keeps that file's directory when the module uses it.

**Migration:** Relative paths written in local module resources or outputs now use the module file's directory. Change paths that relied on the parent template's directory. For example, use `../dist` in `modules/worker.yaml` to refer to `dist` beside the parent template. Parent-supplied properties and overrides, absolute paths, top-level resources, and local asset paths in remote or zip modules keep their existing directories. There is no fallback to the old location.

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
