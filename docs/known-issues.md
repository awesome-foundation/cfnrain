# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead, and links to its issue. When a fix ships, its entry goes away.

## fmt loses or moves comments on a function inside a list { #fmt-drops-a-line-comment-on-a-function-inside-a-list }

When a list item is a function with sequence arguments, `fmt` drops a line comment on the function key. Examples include `Fn::Join` in a list and `Fn::Split` inside `Fn::Select`. On the first run, `fmt` moves a foot comment after a function list item into the next item. A second run changes the output again. Put comments above the property that contains the list instead. See [#127](https://github.com/awesome-foundation/cfnrain/issues/127).

## pkg leaves Rain constants in sequence-form Sub unchanged

`pkg` does not replace `${Rain::X}` in `!Sub [string, {vars}]`. CloudFormation receives the unresolved constant. Use scalar-form `!Sub` for Rain constants, or replace the constant with its value before packaging. See [#126](https://github.com/awesome-foundation/cfnrain/issues/126).

## Modules resources resolve Run and Path from the top-level template

Resources from a `Modules:` section resolve `!Rain::S3` `Run` and `Path` relative to the top-level template directory. They do not use the module file's directory. Use `Type: !Rain::Module` with `--experimental` to resolve paths from the module file's directory instead. See [#89](https://github.com/awesome-foundation/cfnrain/issues/89).

## The upload cache ignores Extension

Two `!Rain::S3` directives with the same `Path` and different `Extension` values share the first upload result. The second directive gets the first extension. Use separate file paths when uploads need different extensions. See [#91](https://github.com/awesome-foundation/cfnrain/issues/91).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`. Issue: [#41](https://github.com/awesome-foundation/cfnrain/issues/41).

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
