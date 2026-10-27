# CI usage

cfnrain works in CI without changes. This page shows how to install it in GitHub Actions, how to deploy without questions, and how to show a change set preview on a pull request.

## Install in GitHub Actions

Use the [install-cfnrain](install.md#github-actions) action in each job that runs cfnrain. The workflows on this page set the version once, in `CFNRAIN_VERSION`, and give it to every install step.

## Deploy without questions

In CI there is nobody to answer a prompt. Use `-y`, and give every parameter a value from `--params`, a config file or a `RAIN_VAR_*` variable. See [Without questions](concepts/deploying.md#without-questions).

```sh
cfnrain deploy app.yaml app-prod -y -c config/prod.yaml --no-colour
```

When the deploy fails, cfnrain exits with a status that is not 0, so the CI job fails. See [Exit status](concepts/deploying.md#exit-status). `--no-colour` removes the color codes from the output, so the CI logs are easier to read.

## A complete workflow

This workflow checks the format of the templates on every pull request and deploys on every push to `main`. It uses OpenID Connect to get AWS credentials, so there are no long-lived keys in the repository.

```yaml
name: Deploy

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read
  id-token: write

env:
  CFNRAIN_VERSION: 1.25.2

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: awesome-foundation/install-cfnrain@v1
        with:
          version: ${{ env.CFNRAIN_VERSION }}
      - name: Check template format
        run: cfnrain fmt -v app.yaml

  deploy:
    if: github.event_name == 'push'
    needs: check
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - uses: awesome-foundation/install-cfnrain@v1
        with:
          version: ${{ env.CFNRAIN_VERSION }}
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789012:role/github-deploy
          aws-region: eu-west-1
      - name: Deploy
        run: cfnrain deploy app.yaml app-prod -y -c config/prod.yaml --no-colour
```

## Change set preview on pull requests

A preview shows the reviewers what a pull request will change in the stack, before anybody merges it. Create a change set with `--no-exec`, read it with the AWS CLI, and post it as a comment.

```yaml
  preview:
    if: github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name == github.repository
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - uses: awesome-foundation/install-cfnrain@v1
        with:
          version: ${{ env.CFNRAIN_VERSION }}
      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::123456789012:role/github-preview
          aws-region: eu-west-1
      - name: Create the change set
        id: changeset
        run: |
          name="pr-${{ github.event.number }}-${{ github.run_attempt }}"
          echo "name=$name" >> "$GITHUB_OUTPUT"
          cfnrain deploy app.yaml app-prod "$name" --no-exec -y -c config/prod.yaml --no-colour | tee deploy.log
      - name: Comment on the pull request
        env:
          GH_TOKEN: ${{ github.token }}
          NAME: ${{ steps.changeset.outputs.name }}
        run: |
          {
            echo "### Change set for app-prod"
            echo
            if grep -q "Deploy was skipped" deploy.log; then
              echo "No changes."
            else
              echo '```'
              aws cloudformation describe-change-set \
                --stack-name app-prod --change-set-name "$NAME" \
                --query 'Changes[].ResourceChange.[Action,LogicalResourceId,ResourceType,Replacement]' \
                --output table
              echo '```'
            fi
          } > preview.md
          gh pr comment "${{ github.event.number }}" --body-file preview.md
      - name: Delete the change set
        if: always()
        run: cfnrain rm -c app-prod "${{ steps.changeset.outputs.name }}" -y || true
```

Notes:

- The third argument to `deploy` names the change set, so the next step can find it. Without it, the name is `<stack>-<unix time>`.
- `-y` stops cfnrain from asking for parameters. `--no-exec` already skips the confirmation.
- If the template has no changes, cfnrain prints `Deploy was skipped` and exits with status 0. The workflow checks for that text. CloudFormation keeps the empty change set with the status `FAILED`, and the last step deletes it.
- The `Replacement` column shows `True` when CloudFormation will replace a resource. Look for it in the review.
- Use a role for the preview that can create change sets but cannot execute them.
- The `if` condition skips pull requests from forks. Those runs get no AWS credentials, and they should not.
- For a stack that does not exist yet, the change set leaves an empty stack in `REVIEW_IN_PROGRESS`. The next `cfnrain deploy` deletes it without a question. See [Updates](concepts/deploying.md#updates).

!!! warning "-x is --no-exec on deploy"
    Write the long flag names in CI scripts. See [Experimental features](concepts/experimental.md#-x-means-no-exec-on-deploy).

## Other checks

- `cfnrain fmt -v <file>` fails when a template is not formatted.
- `cfnrain pkg <file> -o packaged.yaml` writes the packaged template, so you can run `cfn-lint` or `cfn-guard` on it.
- `cfnrain forecast -x <file> <stack>` exits with status 1 when a check fails. It is experimental.

!!! danger "Do not run info --creds in CI"
    It prints your secret key, and CI logs keep it. See [`info`](commands/account/info.md).
