# IfParam and conditions

A module can include or leave out resources based on the properties that the parent sets. There are two ways: the `IfParam` and `IfNotParam` metadata, and normal CloudFormation conditions.

## IfParam and IfNotParam

Add `Metadata.Rain.IfParam` or `Metadata.Rain.IfNotParam` to a resource in a module:

- `IfParam: <Parameter>` keeps the resource only when the parent sets that property.
- `IfNotParam: <Parameter>` keeps the resource only when the parent does not set that property.

cfnrain removes the `IfParam` and `IfNotParam` keys from the output. `IfParam` only works in modules.

Use `IfParam` for optional parts of a module, such as permissions for a resource that the parent passes in. Use `IfNotParam` when the parent can bring its own resource instead of the one that the module creates.

The module, `queue-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/docs/queue-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/docs/ifparam-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/ifparam-expect.yaml"
```

`Orders` sets `DeadLetterArn`, so it gets `Queue` with a redrive policy. `Events` does not, so it gets `SimpleQueue`.

## Conditions in modules

A module can have a `Conditions` section. cfnrain evaluates each condition when it packages the template:

- If the condition resolves to false, cfnrain removes the resources, modules and outputs that use it, and picks the false branch of each `Fn::If`.
- If the condition resolves to true, cfnrain keeps them and picks the true branch.
- cfnrain removes values that are `!Ref AWS::NoValue`.
- If cfnrain cannot resolve the condition, for example because it uses `AWS::Region` or a parameter of the parent template, it copies the condition to the parent template. The condition name gets the module name in front, like a logical ID.

### Resolved conditions

The module:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-intrinsics-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-intrinsics-template.yaml"
```

The packaged template. Module `A` has `Foo: bar`, so every condition is true. Module `B` has `Foo: baz`, so every condition is false and none of its buckets are in the output.

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-intrinsics-expect.yaml"
```

### Unresolved conditions

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-unres-template.yaml"
```

The module, `cond-unres-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-unres-module.yaml"
```

The submodule, `cond-unres-submodule.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-unres-submodule.yaml"
```

The packaged template. The conditions depend on the region and on the parent parameter `P`, so they move to the parent with the prefixes `Foo` and `FooSub`.

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/cond-unres-expect.yaml"
```
