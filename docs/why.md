# Why CloudFormation and Rain

This page explains why cfnrain deploys AWS CloudFormation stacks. It compares CloudFormation with Terraform and Pulumi. Then it compares cfnrain with other ways to deploy stacks.

## Why CloudFormation

CloudFormation creates and updates AWS resources from a template. You describe the resources in YAML or JSON. CloudFormation makes the changes in the correct order.

CloudFormation has these benefits:

- CloudFormation charges no service fee for AWS resources. You pay only for the resources that you create.
- Stacks are scoped to the AWS account where their resources are. IAM controls apply to them.
- The CloudFormation console shows deploy history. Anyone with read access to CloudFormation can see it, and you manage no extra tool for it.
- CloudFormation blocks concurrent changes to the same stack. It manages the stack state at no extra cost.
- CloudFormation rolls back a failed update to the last working state.
- Drift detection finds resources that changed outside the template.
- Change sets show the changes before you run them.
- StackSets deploy one template to many accounts and Regions.

### Compared with Terraform

Terraform has a provider for almost any service. Providers cover AWS, other clouds and software-as-a-service products.

CloudFormation can manage external systems too. You write a custom resource that a Lambda function handles. This works, but it is not as native as a Terraform provider.

Terraform state needs locking when two updates can run at once. That can be two people, or one person who deploys from a laptop and from CI. You can store Terraform state in a bucket, but you must set up locking yourself or pay for a service. HashiCorp sells a cloud service for this. Many paid CI/CD products also solve this problem. CloudFormation blocks concurrent updates to a stack at no cost.

Terraform changed its license to the Business Source License in August 2023. The OpenTofu project forked Terraform 1.5.5 under the Mozilla Public License. OpenTofu is a Linux Foundation project.

Choose Terraform when:

- You need a native provider for a SaaS product or another cloud.
- Your team already uses Terraform modules.

Choose CloudFormation when:

- You use only AWS.
- You want the service to coordinate state and block concurrent changes at no cost.
- You want rollback and drift detection without extra setup.

### Compared with Pulumi

Pulumi defines infrastructure in general-purpose languages, such as TypeScript, Python and Go. Pulumi is open source under the Apache 2.0 license. Like Terraform, Pulumi needs a service or extra setup to block concurrent updates.

Choose Pulumi when your team wants loops and functions in a program. Choose CloudFormation when you want a template that AWS evaluates, with no program to run.

### Compared with Ansible, Chef, Puppet and SaltStack

Ansible, Chef, Puppet and SaltStack are configuration management tools. You write their configuration as code, and they make servers match it. Each one can call cloud APIs through modules or plugins, so each can create resources. None of them provisions infrastructure natively. None of them keeps stack state the way CloudFormation does.

Use these tools to configure servers. Use infrastructure-as-code tools to provision the resources that the servers run on, such as CloudFormation, Terraform or Pulumi.

### Comparison at a glance

| | CloudFormation | Terraform | Pulumi |
|---|---|---|---|
| Template language | YAML or JSON | HCL | TypeScript, Python, Go and other languages |
| Resource coverage | AWS resources, plus custom resources for other systems | Providers for AWS, other clouds and SaaS products | Providers for AWS, other clouds and SaaS products |
| Where state lives | AWS, managed for you | A backend that you configure | A Pulumi Cloud service or a backend that you configure |
| Blocks concurrent updates | Yes, on the same stack, at no cost | With a locking backend or a paid service | With a service or extra setup |
| Automatic rollback | Yes | Not built in | Not built in |
| Drift detection | Built in | Through plan and refresh | Through refresh |
| Preview before change | Change sets | `terraform plan` | `pulumi preview` |
| License | AWS service | Business Source License | Apache 2.0 |
| Cost | No charge for AWS resource types | Free CLI, paid cloud service | Free open source, paid cloud service |

### Limits to plan for

- One template can have up to 500 resources. Split a larger system into nested stacks.
- CloudFormation accepts up to 51,200 bytes of template body in a request. Larger templates must be in Amazon S3, up to 1 MB. cfnrain uploads larger templates for you. See [Why cfnrain](#why-cfnrain).

See the [CloudFormation quotas](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/cloudformation-limits.html) for the current values.

## Why cfnrain

cfnrain is a community fork of [AWS CloudFormation Rain](https://github.com/aws-cloudformation/rain). AWS archived Rain on 2026-07-29. cfnrain works as a replacement for Rain v1.24.4. cfnrain is not affiliated with AWS.

cfnrain is a thin layer over CloudFormation. Each deploy runs as a CloudFormation stack operation. cfnrain adds these features:

- Packaging. `pkg` uploads local files to the artifact bucket and updates the template. See [Packaging](concepts/packaging.md).
- Large templates. cfnrain uploads a template above 51,200 bytes to the artifact bucket. You do not upload it yourself. cfnrain rejects a template above 1 MB.
- Template extensions. Directives such as `!Rain::S3`, `!Rain::Embed` and `!Rain::Include` reuse files in a template. `!Rain::Env` and `!Rain::Constant` reuse values. See [Directives](directives/index.md).
- Modules. Modules reuse template parts across stacks. See [Modules](modules/index.md).
- Stack and template commands. These include `diff`, `watch` and `logs`. See the [stack commands](commands/stack/index.md) and [template commands](commands/template/index.md).

### Compared with the AWS CLI

The AWS CLI deploys a template with `aws cloudformation deploy`. The AWS CLI has a `package` command for local files. It does not process Rain's template extensions, and it does not manage modules or stack commands.

### Compared with AWS SAM

AWS SAM adds a shorthand for serverless resources. SAM also has a CLI for local testing. Use SAM when most of your application runs on Lambda and API Gateway, and you want local testing.

cfnrain does not add serverless resource types. It works with CloudFormation templates.

### Compared with the Serverless Framework

The Serverless Framework has a plugin ecosystem. Version 4 requires a login. Organizations with more than $2 million in annual revenue need a paid subscription. Small businesses and nonprofits can get a discount.

cfnrain uses your AWS credentials. It does not need another account or license.

### Compared with AWS CDK

AWS CDK is a powerful tool for advanced infrastructure as code users. It generates CloudFormation templates from code. The abstraction between your code and CloudFormation is leaky. Its gaps cause surprises and errors in the template, and those are hard to trace back to the code. Newcomers often find this frustrating.

cfnrain does not generate templates from code. Use CDK to generate templates, and use cfnrain to package and deploy them.

## Skip cfnrain when

- You do not use AWS.
- You need one tool for multiple clouds.

For the first steps, see [Install](install.md). To move from Rain, see [Migrating from rain](migrating.md).
