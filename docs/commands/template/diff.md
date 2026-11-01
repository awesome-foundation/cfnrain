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

`diff` treats every YAML NaN value as equal to every other NaN value.

Give `-` as `<from>` or `<to>` to read that template from standard input. For example, to compare the version of a template on the `main` branch with your working copy:

```sh
git show main:./template.yaml | cfnrain diff - template.yaml
```

You can give `-` only once. To use a file that has the name `-`, give it as `./-`.
