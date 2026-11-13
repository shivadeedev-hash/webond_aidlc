---
inclusion: fileMatch
fileMatchPattern: "**/*.tf"
---

# Terraform AWS Development and Deployment Best Practices

## General Rules
1. Always run `terraform fmt` before committing code
2. Always run `terraform validate` before applying changes
3. Use remote state with S3 backend and DynamoDB for state locking
4. Set default region to "us-east-1" for all AWS providers

## Security Rules
1. Never hardcode AWS credentials in Terraform files
2. Use IAM roles with least privilege principle
3. Enable versioning on S3 buckets
4. Encrypt sensitive data at rest and in transit

## Resource Management
1. Always tag resources with "Environment", "Project", and "ManagedBy: Terraform"
2. Use modules for reusable infrastructure components
3. Pin module versions to specific releases
4. Use variables for all configurable values