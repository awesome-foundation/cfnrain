# Known issues

These are places where cfnrain does something unexpected. They come from Rain, and cfnrain has not fixed them yet. Each one says what to do instead, and links to its issue. When a fix ships, its entry goes away.

## fmt drops a line comment on a function inside a list

When a list item is a function with sequence arguments, `fmt` drops a line comment on the function key. Examples include `Fn::Join` in a list and `Fn::Split` inside `Fn::Select`. Put the comment above the property that contains the list instead. See [#85](https://github.com/awesome-foundation/cfnrain/issues/85).

## build --only-cache does the same as --no-cache

`cfnrain build --only-cache` should use only the cached schemas. In the code it sets the same switch as `--no-cache`, so it skips the cache. Do not rely on `--only-cache`. Issue: [#41](https://github.com/awesome-foundation/cfnrain/issues/41).

## The artifact bucket rule says 14 days

When cfnrain creates the artifact bucket, it adds a lifecycle rule with the ID `delete after 14 days`. The rule deletes objects after 7 days. See [Objects expire after 7 days](concepts/artifact-bucket.md#objects-expire-after-7-days). Issue: [#42](https://github.com/awesome-foundation/cfnrain/issues/42).

## The stackset deploy help text says the template is optional

`cfnrain stackset deploy --help` says that without a template file, cfnrain adds a new instance to an existing stack set. The code always reads the first argument as the template file, so you must give one. Issue: [#46](https://github.com/awesome-foundation/cfnrain/issues/46).

If you find another difference between this site and the code, please [open an issue](https://github.com/awesome-foundation/cfnrain/issues).
