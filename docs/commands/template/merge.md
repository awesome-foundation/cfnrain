# merge

`cfnrain merge` merges two or more templates into one.

```sh
cfnrain merge bucket.yaml queue.yaml
```

```yaml
Resources:
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub ${AWS::StackName}-data

  Queue:
    Type: AWS::SQS::Queue
```

If two templates use the same logical ID, `merge` stops with an error. `-f` (`--force`) renames the second one instead. It does not update the `Ref`, `GetAtt` and `Sub` references to the renamed item. `--merge-imports` replaces an `Fn::ImportValue` with a direct reference when the export comes from another template in the merge. `-o` (`--output`) writes the result to a file. `--node-style` sets how to write YAML values, as for [`fmt`](fmt.md).
