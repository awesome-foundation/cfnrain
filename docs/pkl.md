# Pkl

[Pkl](https://pkl-lang.org/) is a configuration language from Apple. cfnrain can read templates written in Pkl, and can write templates as Pkl.

## Requirements

cfnrain runs Pkl through the `pkl` command line tool. Install it from the [Pkl website](https://pkl-lang.org/main/current/pkl-cli/index.html) and make sure that it is on your `PATH`.

## Read Pkl templates

`pkg`, `deploy` and `fmt` accept a file with the `.pkl` extension. cfnrain evaluates it to YAML first, and then works with the YAML.

```sh
cfnrain fmt template.pkl
cfnrain deploy template.pkl my-stack
```

`diff`, `merge` and `tree` do not read Pkl files. Run `cfnrain fmt template.pkl > template.yaml` first.

A Pkl template that uses the [cloudformation-pkl](https://github.com/aws-cloudformation/cloudformation-pkl) package:

```pkl
amends "package://github.com/aws-cloudformation/cloudformation-pkl/releases/download/cloudformation@0.1.1/cloudformation@0.1.1#/template.pkl"
import "package://github.com/aws-cloudformation/cloudformation-pkl/releases/download/cloudformation@0.1.1/cloudformation@0.1.1#/cloudformation.pkl" as cfn
import "package://github.com/aws-cloudformation/cloudformation-pkl/releases/download/cloudformation@0.1.1/cloudformation@0.1.1#/aws/s3/bucket.pkl" as bucket

Description = "Create a bucket"

Parameters {
    ["Name"] {
        Type = "String"
        Default = "baz"
    }
}

Resources {
    ["MyBucket"] = new bucket.Bucket {
        BucketName = cfn.Ref("Name")
    }
}
```

The cloudformation-pkl package is an AWS project. cfnrain does not maintain it.

## Write Pkl

`cfnrain fmt -p` converts a YAML or JSON template to Pkl:

```sh
cfnrain fmt -p template.yaml > template.pkl
```

`--pkl-basic` and `--pkl-package` change the Pkl output. See [`fmt`](commands/template/fmt.md).

`cfnrain build --pkl-class <type>` writes a Pkl class for a resource type.

A full example is in [`test/webapp`](https://github.com/awesome-foundation/cfnrain/tree/main/test/webapp), in `webapp.pkl`.
