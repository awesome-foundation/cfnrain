# ForEach

`ForEach` makes one copy of a module for each item in a list. There are two forms: the `ForEach` key on a module, and `Fn::ForEach` in the `Modules` section.

The list must be known when cfnrain packages the template. It can be a comma-separated string, or `!Ref` to a `CommaDelimitedList` parameter with a `Default`. cfnrain uses the default value. It cannot use a value that CloudFormation resolves at deploy time.

## The ForEach key

Add `ForEach` to a module entry. In the module's `Properties`, two placeholders are available:

| Placeholder | Value |
|-------------|-------|
| `$Identifier` | The current list item |
| `$Index` | The position of the item in the list, from 0 |

There is no `$MapValue` placeholder. An old example in the Rain README used it, but the code never supported it.

cfnrain adds the index to the module name: `Data0`, `Data1`, `Data2`. So the resource `Bucket` becomes `Data0Bucket`, `Data1Bucket` and `Data2Bucket`.

The module, `bucket-module.yaml`:

```yaml
--8<-- "cft/pkg/tmpl/docs/bucket-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/docs/foreach-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/docs/foreach-expect.yaml"
```

## Outputs of a loop

| Reference | Result |
|-----------|--------|
| `!GetAtt Data[2].Arn` | The output of one copy, by index |
| `!GetAtt Data[prod].Arn` | The output of one copy, by list item |
| `!GetAtt Data[].Arn` | A list with the output of every copy |
| `${Data[0].Arn}` in a `!Sub` | The output of one copy, in a string |

In another module with the same `ForEach` list, `!GetAtt Data[$Index].Arn` refers to the copy with the same index.

## Fn::ForEach

`Fn::ForEach::<LoopName>` in the `Modules` section uses the syntax of the CloudFormation `AWS::LanguageExtensions` transform. It takes three items: an identifier name, the list, and a map from a module name to a module configuration. Use the identifier as `${Name}` in the properties, or `&{Name}` in the module name.

If the identifier is in the module name, the copies use the list item in their names, such as `BBFoo`. If it is not, cfnrain adds the index, such as `A0Foo`.

The module:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/foreach-module.yaml"
```

The parent template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/foreach-template.yaml"
```

The packaged template:

```yaml
--8<-- "cft/pkg/tmpl/awscli-modules/foreach-expect.yaml"
```

`Fn::ForEach` outside the `Modules` section is not processed by cfnrain. CloudFormation processes it if the template uses the `AWS::LanguageExtensions` transform.
