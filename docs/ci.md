# CI usage

cfnrain works in CI without changes. This page shows how to install it in GitHub Actions, how to deploy without questions, and how to show a change set preview on a pull request.

## Install in GitHub Actions

Pin the version, download the archive, and check it before you install it:

```yaml
- name: Install cfnrain
  env:
    CFNRAIN_VERSION: 1.25.2
    GH_TOKEN: ${{ github.token }}
  run: |
    cd "$RUNNER_TEMP"
    archive="cfnrain_${CFNRAIN_VERSION}_linux_amd64.tar.gz"
    gh release download "v${CFNRAIN_VERSION}" -R awesome-foundation/cfnrain \
      -p "$archive" -p checksums.txt
    sha256sum --ignore-missing -c checksums.txt
    gh attestation verify "$archive" -R awesome-foundation/cfnrain
    tar -xzf "$archive" cfnrain
    sudo install cfnrain /usr/local/bin/cfnrain
    cfnrain --version
```

`sha256sum` fails the step when the archive does not match `checksums.txt`. `gh attestation verify` fails the step when the archive was not built by the cfnrain release workflow. See [Install](install.md) for the details.

## Deploy without questions

In CI there is nobody to answer a prompt. Use `-y`:

```sh
cfnrain deploy app.yaml app-prod -y -c config/prod.yaml
```

With `-y`:

- cfnrain does not ask for parameters. Every parameter needs a value from `--params`, the config file, a `RAIN_VAR_*` variable, a default, or the existing stack. If one has none, the deploy fails with an error.
- cfnrain does not show the change set before it executes it.
- cfnrain creates the artifact bucket if it does not exist.

The exit status is 0 when the stack reaches `CREATE_COMPLETE`, `UPDATE_COMPLETE` or `IMPORT_COMPLETE`, and when there are no changes. Any other result exits with a status that is not 0, so the CI job fails.

`--no-colour` removes the color codes from the output. This makes CI logs easier to read.

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
      - name: Install cfnrain
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          cd "$RUNNER_TEMP"
          archive="cfnrain_${CFNRAIN_VERSION}_linux_amd64.tar.gz"
          gh release download "v${CFNRAIN_VERSION}" -R awesome-foundation/cfnrain \
            -p "$archive" -p checksums.txt
          sha256sum --ignore-missing -c checksums.txt
          tar -xzf "$archive" cfnrain
          sudo install cfnrain /usr/local/bin/cfnrain
      - name: Check template format
        run: cfnrain fmt -v app.yaml

  deploy:
    if: github.event_name == 'push'
    needs: check
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Install cfnrain
        env:
          GH_TOKEN: ${{ github.token }}
        run: |
          cd "$RUNNER_TEMP"
          archive="cfnrain_${CFNRAIN_VERSION}_linux_amd64.tar.gz"
          gh release download "v${CFNRAIN_VERSION}" -R awesome-foundation/cfnrain \
            -p "$archive" -p checksums.txt
          sha256sum --ignore-missing -c checksums.txt
          tar -xzf "$archive" cfnrain
          sudo install cfnrain /usr/local/bin/cfnrain
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
      # Install cfnrain as in the deploy job
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
- For a stack that does not exist yet, the change set leaves an empty stack in `REVIEW_IN_PROGRESS`. The next `cfnrain deploy` deletes it without a question. See [Getting started](getting-started.md#change-the-stack).

!!! warning "-x is --no-exec on deploy"
    On `deploy`, `-x` means `--no-exec`. On other commands, `-x` means `--experimental`. Write the long flag names in CI scripts, so the intent is clear.

## Other checks

- `cfnrain fmt -v <file>` fails when a template is not formatted.
- `cfnrain pkg <file> -o packaged.yaml` writes the packaged template, so you can run `cfn-lint` or `cfn-guard` on it.
- `cfnrain forecast -x <file> <stack>` exits with status 1 when a check fails. It is experimental.

!!! danger "Do not run info --creds in CI"
    `cfnrain info --creds` prints the secret key and session token. CI logs keep them.
