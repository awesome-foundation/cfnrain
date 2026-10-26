# diff

`cfnrain diff <from> <to>` compares two template files.

```sh
cfnrain diff old.yaml new.yaml
```

```text
(|) Resources:
(|)   Bucket:
(|)     Properties:
(|)       BucketName:
(>)         Fn::Sub: ${AWS::StackName}-files
(+)   Queue:
(+)     Type: AWS::SQS::Queue
```

`(+)` is an addition, `(-)` is a removal, `(>)` is a change and `(|)` is a parent of a change. `-l` (`--long`) also shows the parts that did not change. `diff` compares two local files. It does not compare a file with a deployed stack.
