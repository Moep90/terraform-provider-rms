---
name: Bug Report
description: Create a report to help us improve
title: "[Bug]: "
labels: ["bug", "triage"]
assignees: []
---

## Description

A clear and concise description of what the bug is.

## Terraform/OpenTofu Version

Run `terraform -v` or `tofu -v` and paste the output here.

```
# Paste version output here
```

## Provider Version

What version of the Teltonika RMS provider are you using?

## Affected Resource(s)

Please list the resources affected, e.g.:

- `rms_company`
- `rms_device`
- `rms_tag`
- etc.

## Terraform Configuration Files

```hcl
# Paste your Terraform configuration here
# Please remove any sensitive information (tokens, passwords, etc.)
```

## Debug Output

Please provide debug output if available. You can enable debug logging with:

```bash
TF_LOG=DEBUG terraform apply
```

## Panic Output

If Terraform produced a panic, please provide the full output.

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Steps to Reproduce

1. `terraform init`
2. `terraform apply`
3. ...

## Important Facts

- Operating System and version:
- Go version:
- Any relevant API endpoints or endpoints used:

## Additional Context

Add any other context about the problem here.

## Possible Solution

If you have a suggested solution, please describe it here.
