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

## How it works

### The Modules section

Each entry in `Modules` has a name and a configuration:

| Key | What it does |
|-----|--------------|
| `Source` | Where the module file is. See [Sources](#sources). |
| `Properties` | Values for the module's `Parameters` |
| `Overrides` | Changes to the module's resources. See [Overrides](#overrides). |
| `ForEach` | Make one copy of the module for each item in a list. See [ForEach](foreach.md). |

### Module files

A module looks like a template. It can have `Parameters`, `Resources`, `Conditions`, `Outputs`, `Constants`, `Modules` and `Packages`. Any valid CloudFormation template works as a module.

The parent sets the module's parameters with `Properties`. A property can be a string, a list or an object. Inside the module, `!Ref <Parameter>` and `${Parameter}` in a `!Sub` get the property value. If the parent does not set a property, cfnrain uses the parameter's `Default`.

### Logical IDs

cfnrain adds the module name in front of each logical ID. The resource `Bucket` in the module `Logs` becomes `LogsBucket`. cfnrain also updates the `Ref`, `GetAtt`, `Sub` and `DependsOn` references inside the module.

### Outputs

The parent can use a module output with `!GetAtt <Module>.<Output>` or `${<Module>.<Output>}` in a `!Sub`. cfnrain replaces the reference with the output value. In the example, `!GetAtt Logs.Arn` becomes `!GetAtt LogsBucket.Arn`.

You can also refer to a module resource by its new logical ID, such as `!Ref LogsBucket`. This works, but it breaks when the module author renames the resource. Outputs are safer.

### Overrides

`Overrides` changes the module's resources. The keys are the logical IDs inside the module. cfnrain merges the override into the resource:

- A map merges with the map in the module, key by key.
- A scalar replaces the value in the module.
- A list adds its items to the list in the module. For example, an override can add a statement to a policy.
- This works for `Properties`, `Metadata`, `CreationPolicy` and `UpdatePolicy`. You can also set `DeletionPolicy`, `UpdateReplacePolicy` and `Condition`, and add to `DependsOn`.

There is no way to remove a list item or a property with an override. An override must name a resource that exists in the module. It cannot set a property that has the same name as a module parameter. Set the parameter with `Properties` instead.

Overrides depend on the inside of the module. If the module changes, your overrides can break. Use parameters when the module offers them, and use overrides when it does not.

### Sources

`Source` can be:

- A path relative to the file that has the `Modules` section, such as `./bucket-module.yaml` or `modules/vpc.yaml`. It is not relative to the directory where you run cfnrain.
- An `https://` URL or an `s3://` URI.
- A path inside a package, such as `$samples/encrypted-bucket.yaml`. See [Packages](packages.md).

A module can include other modules. A module cannot include itself, directly or through other modules.

### Conditions

cfnrain evaluates the module's conditions when it can. See [IfParam and conditions](conditions.md).

## Experimental status

The `Modules` section does not need `--experimental`. Rain's old README said that it did, but the code does not check for it. Only the older [`!Rain::Module`](packages.md#the-older-rainmodule-directive) directive needs `--experimental`.

The feature is still young. The format can change in a minor release.

## More tested examples

The module tests in [`cft/pkg/tmpl/awscli-modules`](https://github.com/awesome-foundation/cfnrain/tree/main/cft/pkg/tmpl/awscli-modules) cover more cases. Each `-template.yaml` file has a matching `-expect.yaml` file with the packaged result. This one shows a module output in a `!Sub`, a `GetAtt` inside the module, and an override.

The module:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/basic-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/basic-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/basic-expect.yaml"
```
